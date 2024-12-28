package iplugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/router/chain"
	"github.com/alx99/ika/internal/teardown"
	"github.com/alx99/ika/plugin"
)

type PluginCache struct {
	factories map[string]plugin.Factory
	plugins   map[string]plugin.Plugin
}

type initializedPlugin[T plugin.Plugin] struct {
	name   string
	plugin T
}

func NewPluginCache(factories map[string]plugin.Factory) *PluginCache {
	return &PluginCache{
		factories: factories,
		plugins:   map[string]plugin.Plugin{},
	}
}

func (ps *PluginCache) getPlugin(ctx context.Context, iCtx plugin.InjectionContext, cfg config.Plugin) (plugin.Plugin, bool, error) {
	plugin, ok := ps.plugins[cfg.Name]
	if ok {
		return plugin, false, nil
	}

	factory, ok := ps.factories[cfg.Name]
	if !ok {
		return nil, false, fmt.Errorf("plugin %q not found", cfg.Name)
	}

	plugin, err := factory.New(ctx, iCtx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
	}

	return plugin, true, nil
}

// UsePlugins sets up plugins and calls the provided function with the set up plugins.
// Plugins are set up in the order they are provided in the config.
func UsePlugins[T plugin.Plugin, V any](ctx context.Context,
	iCtx plugin.InjectionContext,
	cache *PluginCache,
	pluginCfg config.Plugins,
	fn func(t []initializedPlugin[T]) V,
) (V, teardown.TeardownFunc, error) {
	var t T
	var v V
	var plugins []initializedPlugin[T] // plugins that have been set up
	var tder teardown.Teardowner

	for _, cfg := range pluginCfg {
		plugin, setup, err := cache.getPlugin(ctx, iCtx, cfg)
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

		iCtx.Logger = iCtx.Logger.With(slog.String("plugin", cfg.Name))
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

func ChainFromReqModifiers(reqModifiers []initializedPlugin[plugin.RequestModifier]) chain.Chain {
	cons := make([]chain.Constructor, len(reqModifiers))

	for i, rm := range reqModifiers {
		cons[i] = chain.Constructor{
			Name: rm.name,
			MiddlewareFunc: func(next plugin.Handler) plugin.Handler {
				return plugin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					r, err := rm.plugin.ModifyRequest(r)
					if err != nil {
						return err
					}
					return next.ServeHTTP(w, r)
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
