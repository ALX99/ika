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

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/logger"
	"github.com/alx99/ika/internal/router"
	"github.com/alx99/ika/internal/server"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
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

	cfg := config.NewRunOpts()
	for _, opt := range opts {
		opt(&cfg)
	}

	if err := run(ctx, cfg); err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("ika has shut down")
		slog.Info("Bye <3")
	}
}

func run(ctx context.Context, opts config.RunOpts) error {
	cfg, err := config.Read(*configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	flush := logger.Initialize(cfg.Ika.Logger)
	defer flush()

	err = cfg.SetRuntimeOpts(opts)
	if err != nil {
		return fmt.Errorf("failed to set runtime options: %w", err)
	}

	router, err := router.MakeRouter(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	s := server.NewServer(router, cfg.Servers)
	err = s.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	slog.Info("ika has started", slog.String("goVersion", runtime.Version()))
	<-ctx.Done()
	slog.Info("Caught shutdown signal, shutting down gracefully...")

	ctx, cancel := context.WithTimeoutCause(
		context.Background(),
		cfg.Ika.GracefulShutdownTimeout,
		fmt.Errorf("could not shut down gracefully in %v", cfg.Ika.GracefulShutdownTimeout),
	)
	defer cancel()

	// Shutdown
	return errors.Join(s.Shutdown(ctx), router.Shutdown(ctx))
}
