package ika

import (
	"context"
	"fmt"

	"github.com/alx99/ika/plugin"
)

type runOpts struct {
	hooks map[string]any
}

// Option represents an option for Run.
type Option func(*runOpts)

// WithHook registers a hook.
func WithHook[T any](name string, hook T) Option {
	return func(cfg *runOpts) {
		if _, ok := any(hook).(plugin.Factory[T]); ok {
			cfg.hooks[name] = hook
			return
		}
		withGeneric(name, noopHookFactory[T]{})(cfg)
	}
}

// withGeneric adds a plugin of any kind
func withGeneric[T any](name string, factory plugin.Factory[T]) Option {
	return func(cfg *runOpts) {
		var t T
		if _, ok := any(t).(plugin.TransportHook); ok {
			cfg.hooks[name] = factory
			return
		}
		if _, ok := any(t).(plugin.FirstHandlerHook); ok {
			cfg.hooks[name] = factory
			return
		}
		if _, ok := any(t).(plugin.MiddlewareHook); ok {
			cfg.hooks[name] = factory
			return
		}

		panic(fmt.Sprintf("register: unknown hook type %T", t))
	}
}

type noopHookFactory[T any] struct{}

func (noopHookFactory[T]) New(_ context.Context) (T, error) {
	var hook T
	return hook, nil
}
