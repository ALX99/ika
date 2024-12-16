package config

import (
	"cmp"
	"strings"
	"time"
)

type Logger struct {
	Level         string        `json:"level" yaml:"level"`
	Format        string        `json:"format" yaml:"format"`
	FlushInterval time.Duration `json:"flushInterval" yaml:"flushInterval"`
	AddSource     bool          `json:"addSource" yaml:"addSource"`
}

func (l *Logger) Normalize() {
	l.Level = strings.ToLower(cmp.Or(l.Level, "info"))
	l.Format = strings.ToLower(cmp.Or(l.Format, "json"))
	l.FlushInterval = cmp.Or(l.FlushInterval, time.Second)
}
