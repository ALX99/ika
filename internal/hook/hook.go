package hook

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	pubHook "github.com/alx99/ika/hook"
	"github.com/alx99/ika/internal/config"
)

type Hook struct {
	// Name of the hook
	Name string
	// Namespaces is a list of namespaces where the hook is enabled.
	Namespaces []string
	// Hook function
	Hook any
}
type Hooks []Hook

// Setup sets up all enabled hooks and returns the hooks and a teardown function.
func Setup(ctx context.Context, hooks map[string]pubHook.Factory, namespaces config.Namespaces) (Hooks, func(context.Context) error, error) {
	var setupHooks Hooks
	teardowns := make(map[string]func(context.Context) error)
	teardownFunc := func(context.Context) error {
		var errs error
		for name, teardown := range teardowns {
			err := teardown(ctx)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to teardown hook %q: %w", name, err))
			}
		}
		return errs
	}

	for _, ns := range namespaces {
		for hookCfg := range ns.Hooks.Enabled() {
			// Try to find the hookFactory
			hookFactory, ok := hooks[hookCfg.Name]
			if !ok {
				return nil, nil, fmt.Errorf("hook %q not found", hookCfg.Name)
			}

			// Try to create a new hook
			hook, err := hookFactory.New(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create hook %q: %w", hookCfg.Name, err)
			}

			// Set up the hook
			err = hook.Setup(ctx, hookCfg.Config)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to setup hook %q: %w", hookCfg.Name, err)
			}
			slog.Debug("Hook set up", "name", hookCfg.Name)
			// Save the teardown function
			teardowns[hookCfg.Name] = hook.Teardown

			added := false
			for _, setupHook := range setupHooks {
				if setupHook.Name == hookCfg.Name {
					setupHook.Namespaces = append(setupHook.Namespaces, ns.Name)
					added = true
				}
			}
			if !added {
				setupHooks = append(setupHooks, Hook{
					Name:       hookCfg.Name,
					Namespaces: []string{ns.Name},
					Hook:       hook,
				})
			}
		}
	}
	return setupHooks, teardownFunc, nil
}

func (hooks Hooks) ApplyTspHooks(ctx context.Context, nsName string, tsp http.RoundTripper) (http.RoundTripper, error) {
	for _, hook := range hooks {
		tspHook, ok := hook.Hook.(pubHook.TransportHook)
		if !ok || !slices.Contains(hook.Namespaces, nsName) {
			continue
		}
		var err error
		tsp, err = tspHook.HookTransport(ctx, tsp)
		if err != nil {
			return nil, fmt.Errorf("failed to apply tsp hook %q: %w", hook.Name, err)
		}
	}
	return tsp, nil
}
