package config

import (
	"iter"
)

type (
	Plugin struct {
		Name    string         `yaml:"name"`
		Enabled Nullable[bool] `yaml:"enabled"`
		Config  map[string]any `yaml:"config"`
	}
	Plugins []Plugin
)

// Enabled returns an iterator that yields all Enabled plugins.
func (h Plugins) Enabled() iter.Seq[Plugin] {
	return func(yield func(Plugin) bool) {
		for _, h := range h {
			if h.Enabled.Or(true) {
				if !yield(h) {
					return
				}
			}
		}
	}
}
