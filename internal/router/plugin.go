package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

func makePlugins[T plugin.Plugin](ctx context.Context,
	iCtx plugin.InjectionContext,
	next http.Handler,
	pluginCfg config.Plugins,
	pFactories map[string]plugin.NFactory,
	makeHandlerFunc func(next http.Handler, t []T) http.Handler,
) (http.Handler, func(context.Context) error, error) {
	// plugins that have been set up
	plugins := []T{}
	initializedPluginIndexes := make(map[string]int)

	var teardowns []func(context.Context) error
	teardown := func(ctx context.Context) error {
		var err error
		for _, t := range teardowns {
			err = errors.Join(err, t(ctx))
		}
		return err
	}

	for _, cfg := range pluginCfg {
		var plugin plugin.Plugin
		if i, ok := initializedPluginIndexes[cfg.Name]; ok {
			plugin = plugins[i]
		} else {
			// Create a new plugin and set it up
			var err error
			plugin, err = pFactories[cfg.Name].New(ctx)
			if err != nil {
				return nil, teardown, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
			}
			teardowns = append(teardowns, plugin.Teardown)

			if err := verifyCapabilities(cfg.Name, plugin, pFactories[cfg.Name].Capabilities()); err != nil {
				return nil, teardown, err
			}

			if !slices.Contains(plugin.InjectionLevels(), iCtx.Level) {
				return nil, teardown, fmt.Errorf("plugin %q can not be injected at the specified level", cfg.Name)
			}

			casted, ok := plugin.(T)
			if !ok {
				return nil, teardown, fmt.Errorf("plugin %q does not implement %T", cfg.Name, casted)
			}

			plugins = append(plugins, casted)
		}

		// NOTE this setup might happen more than once for the same plugin
		err := plugin.Setup(ctx, iCtx, cfg.Config)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to setup plugin %q: %w", cfg.Name, err)
		}
	}

	if len(plugins) == 0 {
		return next, teardown, nil
	}

	return makeHandlerFunc(next, plugins), teardown, nil
}

func handlerFromMiddlewares(next http.Handler, middlewares []plugin.Middleware) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nextE plugin.ErrHandler = plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			next.ServeHTTP(w, r)
			return nil
		})
		for _, middleware := range middlewares {
			var err error
			nextE, err = middleware.Handler(nextE)
			if err != nil {
				panic(err) // todo
			}
		}
		nextE.ServeHTTP(w, r)
	})
}

func handlerFromRequestModifiers(next http.Handler, requestModifiers []plugin.RequestModifier) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		for _, rm := range requestModifiers {
			r, err = rm.ModifyRequest(context.Background(), r)
			if err != nil {
				panic(err) // todo
			}
		}
		next.ServeHTTP(w, r)
	})
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

		}
	}
	return nil
}
