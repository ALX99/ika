package iplugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

type hookFactory struct {
	// name of the hook
	name string
	// namespaces is a list of namespaces where the hook is enabled.
	namespaces []string
	plugin.Factory[any]
}

type hookFactories []hookFactory

type Config struct {
	cfg           config.Config
	hookFactories hookFactories
}

func NewConfig(cfg config.Config) Config {
	return Config{cfg: cfg}
}

func (c *Config) LoadEnabledHooks(ctx context.Context, hooks map[string]any) error {
	var factories hookFactories

	for _, ns := range c.cfg.Namespaces {
		for hookCfg := range ns.Hooks.Enabled() {
			// Try to find the factory
			factory, ok := hooks[hookCfg.Name]
			if !ok {
				return fmt.Errorf("hook %q not found", hookCfg.Name)
			}

			added := false
			for i := range factories {
				if factories[i].name == hookCfg.Name {
					if !slices.Contains(factories[i].namespaces, ns.Name) {
						factories[i].namespaces = append(factories[i].namespaces, ns.Name)
					}
					added = true
				}
			}
			if !added {
				fac, ok := factory.(plugin.Factory[any])
				if !ok {
					return fmt.Errorf("hook %q is not a valid factory", hookCfg.Name)
				}
				factories = append(factories, hookFactory{
					name:       hookCfg.Name,
					namespaces: []string{ns.Name},
					Factory:    fac,
				})
			}
		}
	}
	c.hookFactories = factories

	return nil
}

func (c Config) WrapTransport(ctx context.Context, hooksCfg config.Hooks, tsp http.RoundTripper) (http.RoundTripper, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.TransportHook](ctx, hooksCfg, c.hookFactories)
	if err != nil {
		return nil, teardown, err
	}
	for _, hook := range hooks {
		tsp, err = hook.HookTransport(ctx, tsp)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return tsp, teardown, nil
}

func (c Config) WrapMiddleware(ctx context.Context, hooksCfg config.Hooks, mwName string, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.MiddlewareHook](ctx, hooksCfg, c.hookFactories)
	if err != nil {
		return nil, teardown, err
	}
	for _, hook := range hooks {
		handler, err = hook.HookMiddleware(ctx, mwName, handler)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return handler, teardown, nil
}

func (c Config) WrapFirstHandler(ctx context.Context, hooksCfg config.Hooks, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.FirstHandlerHook](ctx, hooksCfg, c.hookFactories)
	if err != nil {
		return nil, teardown, err
	}
	for _, hook := range hooks {
		handler, err = hook.HookFirstHandler(ctx, handler)
		if err != nil {
			return nil, teardown, fmt.Errorf("failed to apply hook: %w", err)
		}
	}
	return handler, teardown, nil
}

// createHooks creates hooks for the given namespace.
func createHooks[T any](ctx context.Context, hooksCfg config.Hooks, factories hookFactories) ([]T, func(context.Context) error, error) {
	var hooks []T
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

	for hookCfg := range hooksCfg.Enabled() {
		for _, factory := range factories {
			if factory.name != hookCfg.Name {
				continue
			}

			hook, err := factory.New(ctx)
			if err != nil {
				return nil, teardown, fmt.Errorf("failed to create hook %q: %w", factory.name, err)
			}

			if setupHook, ok := hook.(plugin.Setupper); ok {
				err = setupHook.Setup(ctx, hookCfg.Config)
				if err != nil {
					return nil, teardown, fmt.Errorf("failed to setup hook %q: %w", factory.name, err)
				}
			}

			handlerHook, ok := hook.(T)
			if !ok {
				// wrong type, run teardown and continue
				if teardownHook, ok := hook.(plugin.Teardowner); ok {
					if err := teardownHook.Teardown(ctx); err != nil {
						return nil, teardown, nil
					}
				}

				continue
			}
			hooks = append(hooks, handlerHook)
			if teardownHook, ok := hook.(plugin.Teardowner); ok {
				teardowns = append(teardowns, teardownHook.Teardown)
			}
		}
	}
	return hooks, teardown, nil
}
