package config

import (
	"iter"
)

type (
	Hook struct {
		Name    string         `yaml:"name"`
		Enabled Nullable[bool] `yaml:"enabled"`
		Config  map[string]any `yaml:"config"`
	}
	Hooks []Hook
)

// Enabled returns an iterator that yields all Enabled hooks.
func (h Hooks) Enabled() iter.Seq[Hook] {
	return func(yield func(Hook) bool) {
		for _, h := range h {
			if h.Enabled.Or(true) {
				if !yield(h) {
					return
				}
			}
		}
	}
}
