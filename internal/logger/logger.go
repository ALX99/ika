package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/lmittmann/tint"
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
		log = slog.New(slog.NewTextHandler(w, opts))

		// DEBUG override
		if os.Getenv("IKA_DEBUG") != "" && level == slog.LevelDebug {
			log = slog.New(tint.NewHandler(w, &tint.Options{
				Level:     opts.Level,
				AddSource: opts.AddSource,
			}))
		}
	}

	log.Info("Logger initialized", "config", cfg)

	for _, warning := range warnings {
		log.LogAttrs(ctx, slog.LevelWarn, warning)
	}

	if cfg.FlushInterval.Dur().Milliseconds() <= 10 {
		w.SetBuffered(false)
		log.LogAttrs(ctx, slog.LevelDebug, "Log buffering disabled. Flush interval too low")
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
