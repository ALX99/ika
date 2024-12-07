package router

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

func makePluginChain[T plugin.Plugin](ctx context.Context,
	iCtx plugin.InjectionContext,
	pluginCfg config.Plugins,
	pFactories map[string]plugin.NFactory,
	makeHandlerFunc func(t []T) chain.Chain,
) (chain.Chain, func(context.Context) error, error) {
	var ch chain.Chain
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
				return ch, teardown, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
			}
			teardowns = append(teardowns, plugin.Teardown)

			if err := verifyCapabilities(cfg.Name, plugin, pFactories[cfg.Name].Capabilities()); err != nil {
				return ch, teardown, err
			}

			if !slices.Contains(plugin.InjectionLevels(), iCtx.Level) {
				return ch, teardown, fmt.Errorf("plugin %q can not be injected at the specified level", cfg.Name)
			}

			casted, ok := plugin.(T)
			if !ok {
				return ch, teardown, fmt.Errorf("plugin %q does not implement %T", cfg.Name, casted)
			}

			plugins = append(plugins, casted)
		}

		// NOTE this setup might happen more than once for the same plugin
		err := plugin.Setup(ctx, iCtx, cfg.Config)
		if err != nil {
			return ch, teardown, fmt.Errorf("failed to setup plugin %q: %w", cfg.Name, err)
		}
	}

	return makeHandlerFunc(plugins), teardown, nil
}

func handlerFromMiddlewares(middlewares []plugin.Middleware) chain.Chain {
	cons := make([]chain.Constructor, len(middlewares))
	for i := range middlewares {
		cons[i] = middlewares[i].Handler
	}
	return chain.New(cons...)
}

func handlerFromRequestModifiers(requestModifiers []plugin.RequestModifier) chain.Chain {
	fn := func(eh plugin.ErrHandler) plugin.ErrHandler {
		return plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			var err error
			for _, requestModifier := range requestModifiers {
				r, err = requestModifier.ModifyRequest(r.Context(), r)
				if err != nil {
					return err
				}
			}
			return eh.ServeHTTP(w, r)
		})
	}

	return chain.New(fn)
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
