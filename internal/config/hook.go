package config

import (
	"iter"
)

type (
	Plugin struct {
		Name    string         `json:"name"`
		Enabled *bool          `json:"enabled"`
		Config  map[string]any `json:"config"`
	}
	Plugins []Plugin
)

// Enabled returns an iterator that yields all Enabled plugins.
func (p Plugins) Enabled() iter.Seq[Plugin] {
	return func(yield func(Plugin) bool) {
		for _, plugin := range p {
			if plugin.Enabled == nil || *plugin.Enabled {
				if !yield(plugin) {
					return
				}
			}
		}
	}
}

// Names returns an iterator that yields all enabled plugin names.
func (p Plugins) Names() iter.Seq[string] {
	return func(yield func(string) bool) {
		for plugin := range p.Enabled() {
			if !yield(plugin.Name) {
				return
			}
		}
	}
}
