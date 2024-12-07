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

func makePlugins[T any](ctx context.Context,
	iCtx plugin.InjectionContext,
	next http.Handler,
	pluginCfg config.Plugins,
	pFactories map[string]plugin.NFactory,
	makeHandlerFunc func(next http.Handler, t []T) http.Handler,
) (http.Handler, func(context.Context) error, error) {
	var t T
	// plugins that have been set up
	plugins := make(map[string]plugin.Plugin)

	var teardowns []func(context.Context) error
	teardown := func(ctx context.Context) error {
		var err error
		for _, t := range teardowns {
			if e := t(ctx); e != nil {
				err = errors.Join(err, e)
			}
		}
		return err
	}

	for _, cfg := range pluginCfg {
		plugin, ok := plugins[cfg.Name]

		// Create a new plugin and set it up
		if !ok {
			var err error
			plugin, err = pFactories[cfg.Name].New(ctx)
			if err != nil {
				return nil, teardown, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
			}
			teardowns = append(teardowns, plugin.Teardown)

			if !slices.Contains(plugin.InjectionLevels(), iCtx.Level) {
				return nil, teardown, fmt.Errorf("plugin %q can not be injected at the specified level", cfg.Name)
			}

			if _, ok := plugin.(T); !ok {
				return nil, teardown, fmt.Errorf("plugin %q does not implement %T", cfg.Name, t)
			}
			plugins[cfg.Name] = plugin
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

	castedPlugins := make([]T, 0, len(plugins))
	for _, plugins := range plugins {
		castedPlugins = append(castedPlugins, plugins.(T))
	}

	return makeHandlerFunc(next, castedPlugins), teardown, nil
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
