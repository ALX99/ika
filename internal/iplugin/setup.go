package iplugin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/router/chain"
	"github.com/alx99/ika/internal/teardown"
)

type PluginCache struct {
	factories map[string]ika.PluginFactory
	plugins   map[string]ika.Plugin
}

type initializedPlugin[T ika.Plugin] struct {
	name   string
	plugin T
}

func NewPluginCache(factories map[string]ika.PluginFactory) *PluginCache {
	return &PluginCache{
		factories: factories,
		plugins:   map[string]ika.Plugin{},
	}
}

func (ps *PluginCache) getPlugin(ctx context.Context, ictx ika.InjectionContext, cfg config.Plugin) (ika.Plugin, bool, error) {
	plugin, ok := ps.plugins[cfg.Name]
	if ok {
		return plugin, false, nil
	}

	factory, ok := ps.factories[cfg.Name]
	if !ok {
		return nil, false, fmt.Errorf("plugin %q not found", cfg.Name)
	}

	plugin, err := factory.New(ctx, ictx)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
	}

	return plugin, true, nil
}

// UsePlugins sets up plugins and calls the provided function with the set up plugins.
// Plugins are set up in the order they are provided in the config.
func UsePlugins[T ika.Plugin, V any](ctx context.Context,
	ictx ika.InjectionContext,
	cache *PluginCache,
	pluginCfg config.Plugins,
	fn func(t []initializedPlugin[T]) V,
) (V, teardown.TeardownFunc, error) {
	var t T
	var v V
	var plugins []initializedPlugin[T] // plugins that have been set up
	var tder teardown.Teardowner

	for _, cfg := range pluginCfg {
		plugin, setup, err := cache.getPlugin(ctx, ictx, cfg)
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

		ictx.Logger = ictx.Logger.With(slog.String("plugin", cfg.Name))
		// NOTE this setup might happen more than once for the same plugin
		err = plugin.Setup(ctx, ictx, cfg.Config)
		if err != nil {
			return v, tder.Teardown, errors.Join(fmt.Errorf("failed to setup plugin %q: %w", cfg.Name, err), tder.Teardown(ctx))
		}
	}

	return fn(plugins), tder.Teardown, nil
}

func ChainFromMiddlewares(middlewares []initializedPlugin[ika.Middleware]) chain.Chain {
	cons := make([]chain.Constructor, len(middlewares))
	for i := range middlewares {
		cons[i] = chain.Constructor{
			Name:           middlewares[i].name,
			MiddlewareFunc: middlewares[i].plugin.Handler,
		}
	}
	return chain.New(cons...)
}

func ChainFromReqModifiers(reqModifiers []initializedPlugin[ika.RequestModifier]) chain.Chain {
	cons := make([]chain.Constructor, len(reqModifiers))

	for i, rm := range reqModifiers {
		cons[i] = chain.Constructor{
			Name: rm.name,
			MiddlewareFunc: func(next ika.Handler) ika.Handler {
				return ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
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

func MakeTransportWrapper(hooks []initializedPlugin[ika.TripperHooker]) func(http.RoundTripper) (http.RoundTripper, error) {
	var err error
	fn := func(tripper http.RoundTripper) (http.RoundTripper, error) {
		for _, hook := range hooks {
			tripper, err = hook.plugin.HookTripper(tripper)
			if err != nil {
				return nil, err
			}
		}
		return tripper, nil
	}
	return fn
}

func ChainFirstHandler(hooks []initializedPlugin[ika.FirstHandlerHooker]) chain.Chain {
	cons := make([]chain.Constructor, len(hooks))
	for i := range hooks {
		cons[i] = chain.Constructor{
			Name:           hooks[i].name,
			MiddlewareFunc: hooks[i].plugin.Handler,
		}
	}
	return chain.New(cons...)
}
