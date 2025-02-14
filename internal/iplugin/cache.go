package iplugin

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/router/chain"
	"github.com/alx99/ika/internal/teardown"
)

type PluginCache struct {
	factories  map[string]ika.PluginFactory
	teardowner teardown.Teardowner
}

type setuppedPlugin struct {
	name   string
	plugin ika.Plugin
}

func NewPluginCache(factories map[string]ika.PluginFactory) *PluginCache {
	return &PluginCache{factories: factories}
}

func (ps *PluginCache) Teardown(ctx context.Context) error {
	return ps.teardowner.Teardown(ctx)
}

func (ps *PluginCache) GetPlugins(ctx context.Context, ictx ika.InjectionContext, cfgs config.Plugins) ([]setuppedPlugin, error) {
	plugins := make([]setuppedPlugin, 0, len(cfgs))

	for _, cfg := range cfgs {
		ictx.Logger = ictx.Logger.With(slog.String("plugin", cfg.Name))

		factory, ok := ps.factories[cfg.Name]
		if !ok {
			return nil, fmt.Errorf("plugin %q not found", cfg.Name)
		}

		plugin, err := factory.New(ctx, ictx, cfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
		}

		ps.teardowner.Add(plugin.Teardown)

		plugins = append(plugins, setuppedPlugin{
			name:   cfg.Name,
			plugin: plugin,
		})
	}

	return plugins, nil
}

func ChainFromMiddlewares(middlewares []setuppedPlugin) (chain.Chain, error) {
	cons := make([]chain.Constructor, len(middlewares))
	for i := range middlewares {
		mw, ok := middlewares[i].plugin.(ika.Middleware)
		if !ok {
			return chain.Chain{}, fmt.Errorf("plugin %q is not a middleware", middlewares[i].name)
		}

		cons[i] = chain.Constructor{
			Name:           middlewares[i].name,
			MiddlewareFunc: mw.Handler,
		}
	}
	return chain.New(cons...), nil
}

func ChainFromReqModifiers(reqModifiers []setuppedPlugin) (chain.Chain, error) {
	cons := make([]chain.Constructor, len(reqModifiers))

	for i, rm := range reqModifiers {
		modifier, ok := rm.plugin.(ika.RequestModifier)
		if !ok {
			return chain.Chain{}, fmt.Errorf("plugin %q is not a RequestModifier", rm.name)
		}
		cons[i] = chain.Constructor{
			Name: rm.name,
			MiddlewareFunc: func(next ika.Handler) ika.Handler {
				return ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					if err := modifier.ModifyRequest(r); err != nil {
						return err
					}
					return next.ServeHTTP(w, r)
				})
			},
		}
	}

	return chain.New(cons...), nil
}

func MakeTripperHooks(hooks []setuppedPlugin) func(http.RoundTripper) (http.RoundTripper, error) {
	return func(tripper http.RoundTripper) (http.RoundTripper, error) {
		for _, hook := range hooks {
			hooker, ok := hook.plugin.(ika.TripperHooker)
			if !ok {
				continue // hooks does not have to implement every interface
			}
			var err error
			tripper, err = hooker.HookTripper(tripper)
			if err != nil {
				return nil, err
			}
		}
		return tripper, nil
	}
}

func MakeOnRequestHooks(hooks []setuppedPlugin) (chain.Chain, error) {
	cons := make([]chain.Constructor, len(hooks))
	for i, hook := range hooks {
		hooker, ok := hook.plugin.(ika.OnRequestHooker)
		if !ok {
			continue // hooks does not have to implement every interface
		}
		cons[i] = chain.Constructor{
			Name:           hook.name,
			MiddlewareFunc: hooker.Handler,
		}
	}
	return chain.New(cons...), nil
}
