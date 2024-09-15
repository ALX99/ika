package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/hook"
	"github.com/alx99/ika/middleware"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	chimw "github.com/go-chi/chi/v5/middleware"
)

var _ hook.TransportHook = &tracer{}

func init() {
	err := middleware.RegisterFunc("noCache", chimw.NoCache)
	if err != nil {
		panic(err)
	}
}

func main() {
	defer setupMonitoring()()
	ika.Run(ika.WithHook("tracer", &tracer{}))
}

type tracer struct{}

func (w *tracer) Setup(_ context.Context, config map[string]any) error {
	return nil
}

func (w *tracer) Teardown(context.Context) error {
	return nil
}

func (w *tracer) HookTransport(_ context.Context, tsp http.RoundTripper) (http.RoundTripper, error) {
	return otelhttp.NewTransport(tsp), nil
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

	tracerProvider := trace.NewTracerProvider(trace.WithBatcher(traceExporter))
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

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		log.Fatal(err)
	}

	return
}
