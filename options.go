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
func WithHook(name string, hook hook.Hook) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = noopHookFactory{hook}
	}
}

// WithHookFactory registers a hook factory.
func WithHookFactory(name string, factory hook.Factory) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = factory
	}
}

type noopHookFactory struct {
	hook.Hook
}

func (fac noopHookFactory) New(_ context.Context) (hook.Hook, error) {
	return fac, nil
}
