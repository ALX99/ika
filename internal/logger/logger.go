package logger

import (
	"cmp"
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/golang-cz/devslog"
)

type Config struct {
	Level         string        `json:"level" yaml:"level"`
	Format        string        `json:"format" yaml:"format"`
	FlushInterval time.Duration `json:"flushInterval" yaml:"flushInterval"`
	AddSource     bool          `json:"addSource" yaml:"addSource"`
}

func (c *Config) Normalize() {
	c.Level = strings.ToLower(cmp.Or(c.Level, "info"))
	c.Format = strings.ToLower(cmp.Or(c.Format, "json"))
	c.FlushInterval = cmp.Or(c.FlushInterval, time.Second)
}

func Initialize(ctx context.Context, cfg Config) func() error {
	cfg.Normalize()
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
			log = slog.New(devslog.NewHandler(w, &devslog.Options{HandlerOptions: opts}))
		}
	}

	slog.SetDefault(log)
	slog.Info("Logger initialized", "config", cfg)

	for _, warning := range warnings {
		slog.LogAttrs(ctx, slog.LevelWarn, warning)
	}

	go func() {
		t := time.NewTicker(cfg.FlushInterval)
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
	return w.Flush
}
