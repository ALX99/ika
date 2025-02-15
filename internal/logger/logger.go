package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
)

func Initialize(ctx context.Context, cfg config.Logger) (*slog.Logger, func() error) {
	cfg.SetDefaults()
	w := newBufferedWriter(os.Stdout)
	var log *slog.Logger
	var level slog.Level
	var warnings []string

	switch cfg.Level {
	default:
		warnings = append(warnings, "Invalid log level, defaulting to info")
		fallthrough
	case "info":
		level = slog.LevelInfo
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	switch cfg.Format {
	default:
		warnings = append(warnings, "Invalid log format, defaulting to json")
		fallthrough
	case "json":
		log = slog.New(slog.NewJSONHandler(w, opts))
	case "text":
		log = slog.New(tint.NewHandler(w, &tint.Options{
			NoColor:   !isatty.IsTerminal(os.Stdout.Fd()),
			Level:     opts.Level,
			AddSource: opts.AddSource,
		}))
	}

	log.Info("Logger initialized", "config", cfg)

	for _, warning := range warnings {
		log.LogAttrs(ctx, slog.LevelWarn, warning)
	}

	if cfg.FlushInterval.Dur() <= 0 {
		w.SetBuffered(false)
		if err := w.Flush(); err != nil {
			log.Error("Failed to flush logs", "error", err)
		}
		log.LogAttrs(ctx, slog.LevelDebug, "Log buffering disabled")
		return log, func() error { return nil }
	}

	go func() {
		t := time.NewTicker(cfg.FlushInterval.Dur())
		for {
			select {
			case <-ctx.Done():
				w.SetBuffered(false)
				log.LogAttrs(ctx, slog.LevelDebug, "Log buffering disabled")
				t.Stop()
				w.Flush()
				return
			case <-t.C:
				w.Flush()
			}
		}
	}()
	return log, w.Flush
}
