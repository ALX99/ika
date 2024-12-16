package ika

import (
	"fmt"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

// Option represents an option for Run.
type Option func(*config.Options) error

// WithPlugin registers a plugin.
func WithPlugin(p plugin.Factory) Option {
	return func(cfg *config.Options) error {
		if cfg.Plugins == nil {
			cfg.Plugins = make(map[string]plugin.Factory)
		}
		if _, ok := cfg.Plugins[p.Name()]; ok {
			return fmt.Errorf("plugin %q already registered", p.Name())
		}
		cfg.Plugins[p.Name()] = p
		return nil
	}
}
