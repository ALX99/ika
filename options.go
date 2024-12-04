package ika

import (
	"context"
	"fmt"
	"reflect"

	"github.com/alx99/ika/internal/config"
	pplugin "github.com/alx99/ika/plugin"
)

// Option represents an option for Run.
type Option func(*config.RunOpts)

// WithPlugin registers a plugin.
func WithPlugin[T any](name string, plugin T) Option {
	return func(cfg *config.RunOpts) {
		var t T
		if fac, ok := any(plugin).(pplugin.Factory); ok {
			cfg.Plugins[name] = config.PluginFactory{
				PluginVal: reflect.ValueOf(t),
				Factory:   fac,
			}
			return
		}
		withGeneric[T](name, noopPluginFactory[T]{})(cfg)
	}
}

func WithPlugin2(plugin pplugin.NFactory) Option {
	return func(cfg *config.RunOpts) {
		cfg.Plugins2 = append(cfg.Plugins2, plugin)
	}
}

// withGeneric adds a plugin of any kind
func withGeneric[T any](name string, factory pplugin.Factory) Option {
	return func(cfg *config.RunOpts) {
		var t T
		if _, ok := any(t).(pplugin.TransportHook); ok {
			cfg.Plugins[name] = config.PluginFactory{
				PluginVal: reflect.ValueOf(t),
				Factory:   factory,
			}
			return
		}
		if _, ok := any(t).(pplugin.FirstHandlerHook); ok {
			cfg.Plugins[name] = config.PluginFactory{
				PluginVal: reflect.ValueOf(t),
				Factory:   factory,
			}
			return
		}
		if _, ok := any(t).(pplugin.MiddlewareHook); ok {
			cfg.Plugins[name] = config.PluginFactory{
				PluginVal: reflect.ValueOf(t),
				Factory:   factory,
			}
			return
		}

		panic(fmt.Sprintf("register: unknown plugin type %T", factory))
	}
}

type noopPluginFactory[T any] struct{}

var _ pplugin.Factory = noopPluginFactory[any]{}

func (noopPluginFactory[T]) New(_ context.Context) (any, error) {
	var plugin T
	return plugin, nil
}
