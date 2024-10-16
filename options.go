package ika

import (
	"context"
	"fmt"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

// Option represents an option for Run.
type Option func(*config.RunOpts)

// WithHook registers a hook.
func WithHook[T any](name string, hook T) Option {
	return func(cfg *config.RunOpts) {
		if fac, ok := any(hook).(plugin.Factory); ok {
			cfg.Hooks[name] = fac
			return
		}
		withGeneric[T](name, noopPluginFactory[T]{})(cfg)
	}
}

// withGeneric adds a plugin of any kind
func withGeneric[T any](name string, factory plugin.Factory) Option {
	return func(cfg *config.RunOpts) {
		var t T
		if _, ok := any(t).(plugin.TransportHook); ok {
			cfg.Hooks[name] = factory
			return
		}
		if _, ok := any(t).(plugin.FirstHandlerHook); ok {
			cfg.Hooks[name] = factory
			return
		}
		if _, ok := any(t).(plugin.MiddlewareHook); ok {
			cfg.Hooks[name] = factory
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
