package ika

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/router"
	"github.com/alx99/ika/internal/http/server"
	"github.com/alx99/ika/internal/logger"
)

var start = time.Now()

// Run starts Ika
func Run(configPath string, options config.Options) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Read(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	makeServer := func(handler http.Handler, servers []config.Server) server.HTTPServer {
		if options.Validate {
			return &mockServer{onListenAndServe: func() { cancel() }}
		}
		return server.New(handler, servers)
	}

	exitOne := false

	flush, err := run(ctx, makeServer, cfg, options)
	if err != nil {
		slog.Error(err.Error())
		exitOne = !errors.Is(err, context.Canceled)
	} else {
		slog.Info("Shutdown finished")
	}

	if err = flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush: %s\n", err)
	}

	if exitOne || err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context,
	makeServer func(handler http.Handler, servers []config.Server) server.HTTPServer,
	cfg config.Config,
	opts config.Options,
) (func() error, error) {
	log, flush := logger.Initialize(ctx, cfg.Ika.Logger)

	router, err := router.New(cfg, opts, log)
	if err != nil {
		return flush, fmt.Errorf("failed to create router: %w", err)
	}

	err = router.Build(ctx)
	if err != nil {
		return flush, fmt.Errorf("failed to build router: %w", err)
	}

	s := makeServer(router, cfg.Servers)
	err = s.ListenAndServe()
	if err != nil {
		return flush, fmt.Errorf("failed to start: %w", err)
	}

	attrs := []any{
		slog.String("startupTime", time.Since(start).Round(time.Millisecond).String()),
	}

	info, ok := debug.ReadBuildInfo()
	if ok {
		attrs = append(attrs,
			slog.String("version", info.Main.Version),
			slog.String("goVersion", info.GoVersion),
		)
	}
	log.Info("Ika has started", attrs...)

	<-ctx.Done()
	slog.Info("Caught shutdown signal, shutting down gracefully...")

	ctx, cancel := context.WithTimeoutCause(
		context.Background(),
		cfg.Ika.GracefulShutdownTimeout.Dur(),
		fmt.Errorf("could not shut down gracefully in %v", cfg.Ika.GracefulShutdownTimeout),
	)
	defer cancel()

	// Shutdown
	return flush, errors.Join(context.Cause(ctx), s.Shutdown(ctx), router.Shutdown(ctx))
}

func readConfig() (config.Config, error) {
	cfg := config.Config{}
	wd, err := os.Getwd()
	if err != nil {
		return cfg, err
	}

	var cfgFile string
	// look for ika.yaml and ika.json prioritizing json
	for _, ext := range []string{".json", ".yaml"} {
		cfgFile = path.Join(wd, "ika"+ext)
		if _, err := os.Stat(cfgFile); err == nil {
			break
		}
	}

	if cfgFile == "" {
		return cfg, errors.New("ika.json or ika.yaml not found")
	}

	cfg, err = config.Read(cfgFile)
	if err != nil {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}
	return cfg, nil
}

type mockServer struct {
	onListenAndServe func()
}

func (m *mockServer) ListenAndServe() error              { m.onListenAndServe(); return nil }
func (m *mockServer) Shutdown(ctx context.Context) error { return nil }
