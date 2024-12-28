package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/cmd/option"
	"github.com/alx99/ika/plugin"
	"github.com/alx99/ika/plugins"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdkTrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	chimw "github.com/go-chi/chi/v5/middleware"
)

var (
	version                           = "unknown"
	_       plugin.TransportHooker    = &tracer{}
	_       plugin.MiddlewareHook     = &tracer{}
	_       plugin.FirstHandlerHooker = &tracer{}
)

var _ plugin.Middleware = &noCache{}

type noCache struct{}

func (w *noCache) New(context.Context, plugin.InjectionContext) (plugin.Plugin, error) {
	return &noCache{}, nil
}

func (w *noCache) Name() string {
	return "noCache"
}

func (w *noCache) Setup(_ context.Context, _ plugin.InjectionContext, config map[string]any) error {
	return nil
}

func (w *noCache) Teardown(context.Context) error {
	return nil
}

func (w *noCache) Handler(next plugin.Handler) plugin.Handler {
	return plugin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		chimw.NoCache(plugin.ToHTTPHandler(next, nil)).ServeHTTP(w, r)
		return nil
	})
}

func main() {
	defer setupMonitoring()()
	ika.Run(
		option.WithPlugin(&noCache{}),
		option.WithPlugin(&tracer{}),
		option.WithPlugin(plugins.AccessLogger{}),
		option.WithPlugin(plugins.ReqModifier{}),
	)
}

type tracer struct {
	ns string
}

var _ plugin.TransportHooker = &tracer{}

func (w *tracer) New(context.Context, plugin.InjectionContext) (plugin.Plugin, error) {
	return &tracer{}, nil
}

func (w *tracer) Setup(_ context.Context, ictx plugin.InjectionContext, config map[string]any) error {
	w.ns = ictx.Namespace
	return nil
}

func (w *tracer) Name() string {
	return "tracer"
}

func (w *tracer) Teardown(context.Context) error {
	return nil
}

func (w *tracer) HookTransport(tsp http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(tsp,
		otelhttp.WithMetricAttributesFn(metaDataAttrs(w.ns)),
	)
}

func (t *tracer) Handler(next plugin.Handler) plugin.Handler {
	newHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attr := metaDataAttrs(t.ns)(r)
		trace.SpanFromContext(r.Context()).
			SetAttributes(attr...)
		labeler, _ := otelhttp.LabelerFromContext(r.Context())
		labeler.Add(attr...)
		next.ServeHTTP(w, r)
	})

	otelHandler := otelhttp.NewHandler(newHandler, "Request",
		otelhttp.WithPublicEndpoint(),
		otelhttp.WithMetricAttributesFn(metaDataAttrs(t.ns)),
	)

	return plugin.FromHTTPHandler(otelHandler)
}

func (w *tracer) HookMiddleware(_ context.Context, name string, next http.Handler) (http.Handler, error) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := trace.
			SpanFromContext(r.Context()).
			TracerProvider().
			Tracer("example.middleware").
			Start(r.Context(), name, trace.WithSpanKind(trace.SpanKindInternal))
		next.ServeHTTP(w, r)
		span.End()
	}), nil
}

func setupMonitoring() func() {
	p, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   "ika-example",
		ServerAddress:     "https://profiles-prod-019.grafana.net",
		BasicAuthUser:     os.Getenv("PYROSCOPE_USER"),
		BasicAuthPassword: os.Getenv("PYROSCOPE_PASSWORD"),
	})
	if err != nil {
		log.Println("failed to start pyroscope", err)
		return func() {}
	}

	// Set up OpenTelemetry.
	otelShutdown, err := setupOTelSDK(context.Background())
	if err != nil {
		log.Println("failed to setup otel")
	}

	return func() {
		p.Flush(true)
		otelShutdown(context.Background())
	}
}

func metaDataAttrs(ns string) func(r *http.Request) []attribute.KeyValue {
	return func(r *http.Request) []attribute.KeyValue {
		return []attribute.KeyValue{
			attribute.String("namespace.name", ns),
			attribute.String("request.pattern", r.Pattern),
		}
	}
}

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	prop := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(prop)

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		handleErr(err)
		return
	}
	tracerProvider := sdkTrace.NewTracerProvider(
		sdkTrace.WithBatcher(traceExporter),
		sdkTrace.WithResource(r),
	)
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter)))
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	err = host.Start(host.WithMeterProvider(meterProvider))
	if err != nil {
		log.Fatal(err)
	}

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}

	return
}
