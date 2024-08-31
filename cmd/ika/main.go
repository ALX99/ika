package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/router"
	"github.com/alx99/ika/internal/server"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
)

func main() {
	flag.Parse()
	if *printVersion {
		fmt.Println("0.0.1")
		os.Exit(0)
	}

	flush := initLogger()
	defer flush()

	exitWithError := func(msg string, err error) {
		slog.Error(msg, "err", err)
		flush()
		os.Exit(1)
	}

	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		exitWithError("failed to read config", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	defer cancel()

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

	handler, err := router.MakeRouter(ctx, cfg.Namespaces)
	if err != nil {
		exitWithError("failed to create router", err)
	}

	s := server.NewServer(handler, cfg.Server)

	err = s.ListenAndServe()
	if err != nil {
		exitWithError("failed to start", err)
	}

	slog.Info("ika has started")
	<-ctx.Done()
	slog.Info("Caught shutdown signal, shutting down gracefully...")

	ctx, cancel = context.WithTimeoutCause(
		context.Background(),
		cfg.GracefulShutdownTimeout,
		fmt.Errorf("could not shut down gracefully in %v", cfg.GracefulShutdownTimeout),
	)

	defer cancel()
	err = s.Shutdown(ctx)
	if err != nil {
		exitWithError("failed to shut down gracefully", err)
	}

	slog.Info("ika has shut down")
	slog.Info("Bye <3")
}

func initLogger() (flush func() error) {
	w := bufio.NewWriter(os.Stdout)
	log := slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}))
	slog.SetDefault(log)

	return w.Flush
}
