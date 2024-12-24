package config

import (
	"cmp"
	"log/slog"
	"strings"
	"time"
)

type Logger struct {
	Level         string   `json:"level" yaml:"level"`
	Format        string   `json:"format" yaml:"format"`
	FlushInterval Duration `json:"flushInterval" yaml:"flushInterval"`
	AddSource     bool     `json:"addSource" yaml:"addSource"`
}

func (l *Logger) Normalize() {
	l.Level = strings.ToLower(cmp.Or(l.Level, "info"))
	l.Format = strings.ToLower(cmp.Or(l.Format, "json"))
	l.FlushInterval = cmp.Or(l.FlushInterval, Duration(time.Second))
}

func (l Logger) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("level", l.Level),
		slog.String("format", l.Format),
		slog.Any("flushInterval", l.FlushInterval.LogValue()),
		slog.Bool("addSource", l.AddSource),
	)
}
