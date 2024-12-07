package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/middleware"
	"github.com/alx99/ika/plugin"
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
	version                         = "unknown"
	_       plugin.TransportHook    = &tracer{}
	_       plugin.MiddlewareHook   = &tracer{}
	_       plugin.FirstHandlerHook = &tracer{}
)

var _ plugin.Middleware = &noCache{}

type noCache struct{}

func (w *noCache) New(context.Context) (plugin.Plugin, error) {
	return &noCache{}, nil
}

func (w *noCache) Name() string {
	return "noCache"
}

func (w *noCache) Capabilities() []plugin.Capability {
	return []plugin.Capability{plugin.CapMiddleware}
}

func (w *noCache) InjectionLevels() []plugin.InjectionLevel {
	return []plugin.InjectionLevel{plugin.LevelPath}
}

func (w *noCache) Setup(_ context.Context, _ plugin.InjectionContext, config map[string]any) error {
	return nil
}

func (w *noCache) Teardown(context.Context) error {
	return nil
}

func (w *noCache) Handler(_ context.Context, next plugin.ErrHandler) (plugin.ErrHandler, error) {
	return plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		chimw.NoCache(plugin.WrapErrHandler(next)).ServeHTTP(w, r)
		return nil
	}), nil
}

func main() {
	defer setupMonitoring()()
	ika.Run(
		ika.WithPlugin("tracer", &tracer{}),
		ika.WithPlugin2(&noCache{}),
	)
}

type tracer struct{}

func (w *tracer) Setup(_ context.Context, config map[string]any) error {
	return nil
}

func (w *tracer) Teardown(context.Context) error {
	return nil
}

func (w *tracer) HookTransport(_ context.Context, tsp http.RoundTripper) (http.RoundTripper, error) {
	return otelhttp.NewTransport(tsp,
		otelhttp.WithMetricAttributesFn(metaDataAttrs),
	), nil
}

func (w *tracer) HookFirstHandler(_ context.Context, handler http.Handler) (http.Handler, error) {
	newHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attr := metaDataAttrs(r)
		trace.SpanFromContext(r.Context()).
			SetAttributes(attr...)
		labeler, _ := otelhttp.LabelerFromContext(r.Context())
		labeler.Add(attr...)
		handler.ServeHTTP(w, r)
	})

	return otelhttp.NewHandler(newHandler, "Request",
			otelhttp.WithPublicEndpoint(),
			otelhttp.WithMetricAttributesFn(metaDataAttrs),
		),
		nil
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

func metaDataAttrs(r *http.Request) []attribute.KeyValue {
	m := middleware.GetMetadata(r.Context())
	return []attribute.KeyValue{
		attribute.String("namespace.name", m.Namespace),
		attribute.String("ika.route", m.Route),
		attribute.String("ika.generated_route", m.GeneratedRoute),
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
