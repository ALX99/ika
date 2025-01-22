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
	"runtime"
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
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Read(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	makeServer := func(handler http.Handler, servers []config.Server) server.HTTPServer {
		return server.New(handler, servers)
	}

	flush, err := run(ctx, makeServer, cfg, options)
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("Shutdown finished")
	}
	exitOne := !errors.Is(err, context.Canceled)

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
	flush := logger.Initialize(ctx, cfg.Ika.Logger)

	router, err := router.New(cfg, opts, slog.Default())
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

	slog.Info("ika has started",
		slog.String("goVersion", runtime.Version()),
		slog.String("startupTime", time.Since(start).Round(time.Millisecond).String()))

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
