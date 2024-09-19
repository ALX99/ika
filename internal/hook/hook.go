package hook

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/plugin"
)

type HookFactory struct {
	// Name of the hook
	Name string
	// Namespaces is a list of namespaces where the hook is enabled.
	Namespaces []string
	plugin.Factory[plugin.Hook]
}
type HookFactories []HookFactory

// GetFactories returns hook factories for all enabled hooks in all namespaces.
func GetFactories(ctx context.Context, hooks map[string]plugin.Factory[plugin.Hook], namespaces config.Namespaces) (HookFactories, error) {
	var factories HookFactories

	for _, ns := range namespaces {
		for hookCfg := range ns.Hooks.Enabled() {
			// Try to find the hookFactory
			hookFactory, ok := hooks[hookCfg.Name]
			if !ok {
				return nil, fmt.Errorf("hook %q not found", hookCfg.Name)
			}

			added := false
			for i := range factories {
				if factories[i].Name == hookCfg.Name {
					if !slices.Contains(factories[i].Namespaces, ns.Name) {
						factories[i].Namespaces = append(factories[i].Namespaces, ns.Name)
					}
					added = true
				}
			}
			if !added {
				factories = append(factories, HookFactory{
					Name:       hookCfg.Name,
					Namespaces: []string{ns.Name},
					Factory:    hookFactory,
				})
			}
		}
	}
	return factories, nil
}

func (hf HookFactories) ApplyTspHooks(ctx context.Context, hooksCfg config.Hooks, tsp http.RoundTripper) (http.RoundTripper, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.TransportHook](ctx, hooksCfg, hf)
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

func (hf HookFactories) ApplyMiddlewareHooks(ctx context.Context, hooksCfg config.Hooks, mwName string, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.MiddlewareHook](ctx, hooksCfg, hf)
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

func (hf HookFactories) ApplyFirstHandlerHook(ctx context.Context, hooksCfg config.Hooks, handler http.Handler) (http.Handler, func(context.Context) error, error) {
	hooks, teardown, err := createHooks[plugin.FirstHandlerHook](ctx, hooksCfg, hf)
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
func createHooks[T any](ctx context.Context, hooksCfg config.Hooks, factories HookFactories) ([]T, func(context.Context) error, error) {
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
			if factory.Name != hookCfg.Name {
				continue
			}

			hook, err := factory.New(ctx)
			if err != nil {
				return nil, teardown, fmt.Errorf("failed to create hook %q: %w", factory.Name, err)
			}

			err = hook.Setup(ctx, hookCfg.Config)
			if err != nil {
				return nil, teardown, fmt.Errorf("failed to setup hook %q: %w", factory.Name, err)
			}
			teardowns = append(teardowns, hook.Teardown)

			handlerHook, ok := hook.(T)
			if !ok {
				continue
			}
			hooks = append(hooks, handlerHook)
		}
	}
	return hooks, teardown, nil
}
