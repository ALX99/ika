package ika

import (
	"context"

	"github.com/alx99/ika/hook"
)

type startCfg struct {
	hooks map[string]hook.Factory
}

// Option represents an option for Run.
type Option func(*startCfg)

// WithHook registers a hook.
func WithHook[T hook.Hook](name string, hook T) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = noopHookFactory[T]{}
	}
}

// WithHookFactory registers a hook factory.
func WithHookFactory(name string, factory hook.Factory) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = factory
	}
}

type noopHookFactory[T hook.Hook] struct{}

func (noopHookFactory[T]) New(_ context.Context) (hook.Hook, error) {
	var hook T
	return hook, nil
}
