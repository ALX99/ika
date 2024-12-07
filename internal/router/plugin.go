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
	pConfig config.Plugins,
	pFactories map[string]plugin.NFactory,
	makeHandlerFunc func(next http.Handler, t []T) http.Handler,
) (http.Handler, func(context.Context) error, error) {
	plugins := []T{}

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

	for _, pluginCfg := range pConfig {
		plugin, err := pFactories[pluginCfg.Name].New(ctx)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to create plugin %q: %w", pluginCfg.Name, err)
		}
		teardowns = append(teardowns, plugin.Teardown)

		if !slices.Contains(plugin.InjectionLevels(), iCtx.Level) {
			return nil, teardown, fmt.Errorf("plugin %q can not be injected at the specified level", plugin.Name())
		}

		err = plugin.Setup(ctx, iCtx, pluginCfg.Config)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to setup plugin %q: %w", plugin.Name(), err)
		}

		castedPlugin, ok := plugin.(T)
		if !ok {
			return nil, teardown, fmt.Errorf("plugin %q does not implement Middleware", plugin.Name())
		}

		plugins = append(plugins, castedPlugin)
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
