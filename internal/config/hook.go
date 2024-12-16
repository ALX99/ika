package config

import (
	"iter"
)

type (
	Plugin struct {
		Name    string         `yaml:"name"`
		Enabled bool           `yaml:"enabled"`
		Config  map[string]any `yaml:"config"`
	}
	Plugins []Plugin
)

// Enabled returns an iterator that yields all Enabled plugins.
func (p Plugins) Enabled() iter.Seq[Plugin] {
	return func(yield func(Plugin) bool) {
		for _, h := range p {
			if true { // FIXME: h.Enabled
				if !yield(h) {
					return
				}
			}
		}
	}
}

// Names returns an iterator that yields all enabled plugin names.
func (p Plugins) Names() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, h := range p {
			if !yield(h.Name) {
				return
			}
		}
	}
}
