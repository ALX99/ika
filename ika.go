package ika

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/logger"
	"github.com/alx99/ika/internal/router"
	"github.com/alx99/ika/internal/server"
	"github.com/alx99/ika/plugin"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
	start        = time.Now()
)

// Run starts Ika
func Run(opts ...Option) {
	flag.Parse()
	if *printVersion {
		fmt.Println("0.0.1")
		os.Exit(0)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()

	cfg := config.Options{
		Plugins: make(map[string]plugin.Factory),
	}

	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to apply option: %s\n", err)
			os.Exit(1)
		}
	}

	flush, err := run(ctx, cfg)
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("ika has shut down")
		slog.Info("Bye <3")
	}

	if err = flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush: %s\n", err)
	}

	if err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, opts config.Options) (func() error, error) {
	cfg, err := config.Read(*configPath)
	if err != nil {
		return func() error { return nil }, fmt.Errorf("failed to read config: %w", err)
	}

	flush := logger.Initialize(ctx, cfg.Ika.Logger)

	router, err := router.MakeRouter(ctx, cfg, opts)
	if err != nil {
		return flush, fmt.Errorf("failed to create router: %w", err)
	}

	s := server.NewServer(router, cfg.Servers)
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
	return flush, errors.Join(s.Shutdown(ctx), router.Shutdown(ctx))
}
