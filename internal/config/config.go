package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"slices"

	pplugin "github.com/alx99/ika/plugin"
	"gopkg.in/yaml.v3"
)

type PluginFactory struct {
	PluginVal reflect.Value
	Factory   pplugin.Factory

	// name of the plugin
	name string
	// namespaces is a list of namespaces where the plugin is enabled
	namespaces []string
}

type RunOpts struct {
	Plugins map[string]PluginFactory
}

type Config struct {
	Server            Server     `yaml:"server"`
	Namespaces        Namespaces `yaml:"namespaces"`
	NamespaceOverride Namespace  `yaml:"namespaceOverrides"`
	Ika               Ika        `yaml:"ika"`

	// runtime configuration
	pluginFactories []PluginFactory
}

func Read(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}

	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	cfg.ApplyOverride()
	return cfg, nil
}

func NewRunOpts() RunOpts {
	return RunOpts{
		Plugins: make(map[string]PluginFactory),
	}
}

func (c *Config) SetRuntimeOpts(opts RunOpts) error {
	if err := c.loadPlugins(opts.Plugins); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadPlugins(factories map[string]PluginFactory) error {
	for _, ns := range c.Namespaces {
		for cfg := range ns.Plugins.Enabled() {
			// Try to find the factory
			factory, ok := factories[cfg.Name]
			if !ok {
				return fmt.Errorf("plugin %q not found", cfg.Name)
			}

			// Update information
			factory.name = cfg.Name
			factory.namespaces = slices.Compact(append(factory.namespaces, ns.Name))
			factories[cfg.Name] = factory
		}
	}

	for _, factory := range factories {
		if factory.namespaces != nil {
			c.pluginFactories = append(c.pluginFactories, factory)
		}
	}

	return nil
}

func (c Config) WrapTransport(ctx context.Context, pluginsCfg Plugins, tsp http.RoundTripper) (http.RoundTripper, func(context.Context) error, error) {
	hooks, teardown, err := createPlugins[pplugin.TransportHook](ctx, pluginsCfg, c.pluginFactories)
	if err != nil {
		return nil, nil, err
	}
	for _, hook := range hooks {
		tsp, err = hook.HookTransport(ctx, tsp)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return tsp, teardown, nil
}

func (c Config) WrapMiddleware(ctx context.Context, hooksCfg Plugins, mwName string, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createPlugins[pplugin.MiddlewareHook](ctx, hooksCfg, c.pluginFactories)
	if err != nil {
		return nil, nil, err
	}
	for _, hook := range hooks {
		handler, err = hook.HookMiddleware(ctx, mwName, handler)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return handler, teardown, nil
}

func (c Config) WrapFirstHandler(ctx context.Context, hooksCfg Plugins, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createPlugins[pplugin.FirstHandlerHook](ctx, hooksCfg, c.pluginFactories)
	if err != nil {
		return nil, teardown, err
	}
	for _, hook := range hooks {
		handler, err = hook.HookFirstHandler(ctx, handler)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return handler, teardown, nil
}

// ApplyOverride applies the NamespaceOverrides to the Config.
func (c *Config) ApplyOverride() {
	for name, ns := range c.Namespaces {
		ns.Transport = override(ns.Transport, c.NamespaceOverride.Transport)
		c.Namespaces[name] = ns
	}
}

// createPlugins creates plugins for the given namespace.
func createPlugins[T any](ctx context.Context, pluginsCfg Plugins, factories []PluginFactory) ([]T, func(context.Context) error, error) {
	var plugins []T
	var teardowns []func(context.Context) error

	teardown := func(ctx context.Context) error {
		var err error
		for _, t := range teardowns {
			err = errors.Join(err, t(ctx))
		}
		return err
	}

	for pluginCfg := range pluginsCfg.Enabled() {
		for _, factory := range factories {
			if factory.name != pluginCfg.Name {
				continue
			}

			if _, ok := factory.PluginVal.Interface().(T); !ok {
				var t T
				return nil, teardown, fmt.Errorf("plugin %q of type %T does not implement %T", factory.name, factory.PluginVal.Interface(), t)
			}

			plugin, err := factory.Factory.New(ctx)
			if err != nil {
				return nil, nil, errors.Join(
					fmt.Errorf("failed to create plugin %q: %w", factory.name, err),
					teardown(ctx),
				)
			}

			if setupper, ok := plugin.(pplugin.Setupper); ok {
				err = setupper.Setup(ctx, pluginCfg.Config)
				if err != nil {
					return nil, nil, errors.Join(
						fmt.Errorf("failed to setup plugin %q: %w", factory.name, err),
						teardown(ctx),
					)
				}
			}

			typedPlugin, ok := plugin.(T)
			if !ok {
				return nil, nil, errors.Join(
					errors.New("developer error: failed to cast plugin"),
					teardown(ctx),
				)
			}
			plugins = append(plugins, typedPlugin)
			if teardowner, ok := plugin.(pplugin.Teardowner); ok {
				teardowns = append(teardowns, teardowner.Teardown)
			}
		}
	}
	return plugins, teardown, nil
}
