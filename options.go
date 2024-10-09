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
		if _, ok := any(hook).(plugin.Factory); ok {
			cfg.hooks[name] = hook
			return
		}
		withGeneric[T](name, noopPluginFactory[T]{})(cfg)
	}
}

// withGeneric adds a plugin of any kind
func withGeneric[T any](name string, factory any) Option {
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

		panic(fmt.Sprintf("register: unknown plugin type %T", factory))
	}
}

type noopPluginFactory[T any] struct{}

var _ plugin.Factory = noopPluginFactory[any]{}

func (noopPluginFactory[T]) New(_ context.Context) (any, error) {
	var plugin T
	return plugin, nil
}
