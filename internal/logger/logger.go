package logger

import (
	"bufio"
	"cmp"
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

func Initialize(cfg Config) func() error {
	cfg.Normalize()
	w := newBufferedWriter(bufio.NewWriterSize(os.Stdout, 32*1024))
	var log *slog.Logger
	var level slog.Level

	switch cfg.Level {
	default:
		// todo log warning
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
		// todo log warning
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

	go func() {
		t := time.NewTicker(cfg.FlushInterval)
		for range t.C {
			w.Flush()
		}
	}()
	return w.Flush
}
