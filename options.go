package ika

import (
	"context"

	"github.com/alx99/ika/plugin"
)

type startCfg struct {
	hooks map[string]plugin.Factory[plugin.Hook]
}

// Option represents an option for Run.
type Option func(*startCfg)

// WithHook registers a hook.
func WithHook[T plugin.Hook](name string, hook T) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = noopHookFactory[T]{}
	}
}

// WithHookFactory registers a hook factory.
func WithHookFactory(name string, factory plugin.Factory[plugin.Hook]) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = factory
	}
}

type noopHookFactory[T plugin.Hook] struct{}

func (noopHookFactory[T]) New(_ context.Context) (plugin.Hook, error) {
	var hook T
	return hook, nil
}
