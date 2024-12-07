package router

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

func makePlugins[T any](ctx context.Context,
	iCtx plugin.InjectionContext,
	next http.Handler,
	plugins config.Plugins,
	pFactories map[string]plugin.NFactory,
	makeHandlerFunc func(next http.Handler, t []T) http.Handler,
) (http.Handler, error) {
	ps := []T{}
	for _, pluginCfg := range plugins {
		p, err := pFactories[pluginCfg.Name].New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin %q: %w", pluginCfg.Name, err)
		}

		if !slices.Contains(p.InjectionLevels(), iCtx.Level) {
			return nil, fmt.Errorf("plugin %q can not be injected at the specified level", p.Name())
		}

		err = p.Setup(ctx, iCtx, pluginCfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup plugin %q: %w", p.Name(), err)
		}

		mw, ok := p.(T)
		if !ok {
			return nil, fmt.Errorf("plugin %q does not implement Middleware", p.Name())
		}

		ps = append(ps, mw)
	}
	if len(ps) == 0 {
		return next, nil
	}
	return makeHandlerFunc(next, ps), nil
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
