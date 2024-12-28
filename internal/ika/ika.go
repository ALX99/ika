package ika

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

	flush, err := run(ctx, configPath, options)
	if err != nil {
		slog.Error(err.Error())
	} else {
		slog.Info("ika has shut down")
		slog.Info("Bye <3")
	}
	exitOne := !errors.Is(err, context.Canceled)

	if err = flush(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to flush: %s\n", err)
	}

	if exitOne || err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, configPath string, opts config.Options) (func() error, error) {
	if configPath == "" {
		var err error
		configPath, err = locateConfig()
		if err != nil {
			return func() error { return nil }, err
		}
	}

	cfg, err := config.Read(configPath)
	if err != nil {
		return func() error { return nil }, fmt.Errorf("failed to read config: %w", err)
	}

	flush := logger.Initialize(ctx, cfg.Ika.Logger)

	router, err := router.New(cfg, opts, slog.Default())
	if err != nil {
		return flush, fmt.Errorf("failed to create router: %w", err)
	}

	err = router.Build(ctx)
	if err != nil {
		return flush, fmt.Errorf("failed to build router: %w", err)
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

func locateConfig() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// look for ika.yaml and ika.json prioritizing json
	for _, ext := range []string{".json", ".yaml"} {
		path := path.Join(wd, "ika"+ext)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("ika.json or ika.yaml not found")
}
