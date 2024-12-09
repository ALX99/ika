package iplugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/router/chain"
	"github.com/alx99/ika/plugin"
)

type PluginSetupper struct {
	factories                map[string]plugin.Factory
	plugins                  []plugin.Plugin
	initializedPluginIndexes map[string]int
}

type initializedPlugin[T plugin.Plugin] struct {
	name   string
	plugin T
}

func NewSetupper(factories map[string]plugin.Factory) *PluginSetupper {
	return &PluginSetupper{
		factories:                factories,
		initializedPluginIndexes: map[string]int{},
	}
}

func (ps *PluginSetupper) getPlugin(ctx context.Context, iCtx plugin.InjectionContext, cfg config.Plugin) (plugin.Plugin, bool, error) {
	key := cfg.Name
	switch iCtx.Level {
	case plugin.LevelPath:
		key += "_" + iCtx.PathPattern
	case plugin.LevelNamespace:
		key += "_" + iCtx.Namespace
	}

	if i, ok := ps.initializedPluginIndexes[key]; ok {
		return ps.plugins[i], false, nil
	}

	factory, ok := ps.factories[cfg.Name]
	if !ok {
		return nil, false, fmt.Errorf("plugin %q not found", cfg.Name)
	}

	// Create a new plugin and set it up
	plugin, err := factory.New(ctx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
	}

	if err := verifyCapabilities(cfg.Name, plugin, ps.factories[cfg.Name].Capabilities()); err != nil {
		return nil, false, err
	}

	if !slices.Contains(plugin.InjectionLevels(), iCtx.Level) {
		return nil, false, fmt.Errorf("plugin %q can not be injected at the specified level", cfg.Name)
	}

	ps.plugins = append(ps.plugins, plugin)
	ps.initializedPluginIndexes[cfg.Name] = len(ps.plugins) - 1

	return plugin, true, nil
}

// UsePlugins sets up plugins and calls the provided function with the set up plugins.
// Plugins are set up in the order they are provided in the config.
//
// If a plugin injected on the same level has already been created, it will be reused.
// This means that:
// - The same plugin injected multiple times on the same path will be reused.
// - The same plugin injected multiple times on the same namespace will be reused.
func UsePlugins[T plugin.Plugin, V any](ctx context.Context,
	iCtx plugin.InjectionContext,
	setupper *PluginSetupper,
	pluginCfg config.Plugins,
	fn func(t []initializedPlugin[T]) V,
) (V, func(context.Context) error, error) {
	var v V
	// plugins that have been set up
	plugins := []initializedPlugin[T]{}

	var teardowns []func(context.Context) error
	teardown := func(ctx context.Context) error {
		var err error
		for _, t := range teardowns {
			err = errors.Join(err, t(ctx))
		}
		return err
	}

	for _, cfg := range pluginCfg {
		plugin, setup, err := setupper.getPlugin(ctx, iCtx, cfg)
		if err != nil {
			return v, teardown, err
		}

		if setup {
			plugins = append(plugins, initializedPlugin[T]{
				name:   cfg.Name,
				plugin: plugin.(T),
			})
			teardowns = append(teardowns, plugin.Teardown)
		}

		// NOTE this setup might happen more than once for the same plugin
		err = plugin.Setup(ctx, iCtx, cfg.Config)
		if err != nil {
			return v, teardown, fmt.Errorf("failed to setup plugin %q: %w", cfg.Name, err)
		}

	}

	return fn(plugins), teardown, nil
}

func ChainFromMiddlewares(middlewares []initializedPlugin[plugin.Middleware]) chain.Chain {
	cons := make([]chain.Constructor, len(middlewares))
	for i := range middlewares {
		cons[i] = chain.Constructor{
			Name:           middlewares[i].name,
			MiddlewareFunc: middlewares[i].plugin.Handler,
		}
	}
	return chain.New(cons...)
}

func ChainFromReqModifiers(requestModifiers []initializedPlugin[plugin.RequestModifier]) chain.Chain {
	cons := make([]chain.Constructor, len(requestModifiers))

	for i, rm := range requestModifiers {
		cons[i] = chain.Constructor{
			Name: rm.name,
			MiddlewareFunc: func(eh plugin.ErrHandler) plugin.ErrHandler {
				return plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					r, err := rm.plugin.ModifyRequest(r)
					if err != nil {
						return err
					}
					return eh.ServeHTTP(w, r)
				})
			},
		}
	}

	return chain.New(cons...)
}

func MakeTransportWrapper(hooks []initializedPlugin[plugin.TransportHooker]) func(http.RoundTripper) http.RoundTripper {
	fn := func(tsp http.RoundTripper) http.RoundTripper {
		for _, hook := range hooks {
			tsp = hook.plugin.HookTransport(tsp)
		}
		return tsp
	}
	return fn
}

func ChainFirstHandler(hooks []initializedPlugin[plugin.FirstHandlerHooker]) chain.Chain {
	cons := make([]chain.Constructor, len(hooks))
	for i := range hooks {
		cons[i] = chain.Constructor{
			Name:           hooks[i].name,
			MiddlewareFunc: hooks[i].plugin.HookFirstHandler,
		}
	}
	return chain.New(cons...)
}

func verifyCapabilities(pluginName string, p plugin.Plugin, capabilities []plugin.Capability) error {
	var t1 plugin.RequestModifier
	var t2 plugin.Middleware
	for _, capability := range capabilities {
		switch capability {
		case plugin.CapModifyRequests:
			if _, ok := p.(plugin.RequestModifier); !ok {
				return fmt.Errorf("plugin %q does not implement %T", pluginName, t1)
			}
		case plugin.CapMiddleware:
			if _, ok := p.(plugin.Middleware); !ok {
				return fmt.Errorf("plugin %q does not implement %T", pluginName, t2)
			}
		case plugin.CapFirstHandler:
			if _, ok := p.(plugin.FirstHandlerHooker); !ok {
				return fmt.Errorf("plugin %q does not implement %T", pluginName, t2)
			}

		}
	}
	return nil
}
