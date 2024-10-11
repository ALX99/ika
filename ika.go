package ika

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/iplugin"
	"github.com/alx99/ika/internal/router"
	"github.com/alx99/ika/internal/server"
	"github.com/lmittmann/tint"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
	logFormat    = flag.String("log-format", "json", "Log format (json or text)")
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

	flush := initLogger()
	defer flush()

	// flush every 5 seconds
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				flush()
			}
		}
	}()

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

	pCfg := iplugin.NewConfig(cfg)

	if err := pCfg.LoadEnabledHooks(ctx, opts.Hooks); err != nil {
		return fmt.Errorf("failed to load hooks: %w", err)
	}

	router, err := router.MakeRouter(ctx, cfg.Namespaces, pCfg)
	if err != nil {
		return fmt.Errorf("failed to create router: %w", err)
	}

	s := server.NewServer(router, cfg.Server)
	err = s.ListenAndServe()
	if err != nil {
		return fmt.Errorf("failed to start: %w", err)
	}

	slog.Info("ika has started")
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

func initLogger() (flush func() error) {
	w := bufio.NewWriter(os.Stdout)
	var log *slog.Logger
	if *logFormat == "text" {
		log = slog.New(tint.NewHandler(w, &tint.Options{
			Level: slog.LevelDebug,
			// AddSource: true,
		}))
	} else {
		log = slog.New(slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			// AddSource: true,
		}))
	}

	slog.SetDefault(log)

	return w.Flush
}
