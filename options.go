package ika

import (
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

// Option represents an option for Run.
type Option func(*config.Options)

// WithPlugin registers a plugin.
func WithPlugin(plugin plugin.Factory) Option {
	return func(cfg *config.Options) {
		cfg.Plugins[plugin.Name()] = plugin
	}
}
