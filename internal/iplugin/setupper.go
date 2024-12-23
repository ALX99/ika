package iplugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/router/chain"
	"github.com/alx99/ika/internal/teardown"
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
	plugin, err := factory.New(ctx, iCtx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
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
) (V, teardown.TeardownFunc, error) {
	var t T
	var v V
	// plugins that have been set up
	plugins := []initializedPlugin[T]{}
	var tder teardown.Teardowner

	for _, cfg := range pluginCfg {
		plugin, setup, err := setupper.getPlugin(ctx, iCtx, cfg)
		if err != nil {
			return v, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
		}

		if setup {
			tder.Add(plugin.Teardown)
			castedPlugin, ok := plugin.(T)
			if !ok {
				return v, tder.Teardown, errors.Join(fmt.Errorf("plugin %q does not implement interface %T", cfg.Name, t), tder.Teardown(ctx))
			}
			plugins = append(plugins, initializedPlugin[T]{
				name:   cfg.Name,
				plugin: castedPlugin,
			})
		}

		// NOTE this setup might happen more than once for the same plugin
		err = plugin.Setup(ctx, iCtx, cfg.Config)
		if err != nil {
			return v, tder.Teardown, errors.Join(fmt.Errorf("failed to setup plugin %q: %w", cfg.Name, err), tder.Teardown(ctx))
		}
	}

	return fn(plugins), tder.Teardown, nil
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
			MiddlewareFunc: hooks[i].plugin.Handler,
		}
	}
	return chain.New(cons...)
}
