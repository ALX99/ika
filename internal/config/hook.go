package config

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"

	"github.com/alx99/ika/hook"
)

type (
	Hook struct {
		Name    string         `yaml:"name"`
		Enabled Nullable[bool] `yaml:"enabled"`
		Config  map[string]any `yaml:"config"`
	}
	Hooks []Hook
)

// Enabled returns an iterator that yields all Enabled hooks.
func (h Hooks) Enabled() iter.Seq[Hook] {
	return func(yield func(Hook) bool) {
		for _, h := range h {
			if h.Enabled.Or(true) {
				if !yield(h) {
					return
				}
			}
		}
	}
}

// Setup sets up all enabled hooks and returns a teardown function.
func (h Hooks) Setup(ctx context.Context, hooks map[string]hook.Factory) (func(context.Context) error, error) {
	teardowns := make(map[string]func(context.Context) error, len(h))
	teardownFunc := func(ctx context.Context) error {
		var errs error
		for name, teardown := range teardowns {
			err := teardown(ctx)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("failed to teardown hook %q: %w", name, err))
			}
		}
		return errs
	}

	for hookCfg := range h.Enabled() {
		// Try to find the hookFactory
		hookFactory, ok := hooks[hookCfg.Name]
		if !ok {
			return nil, fmt.Errorf("hook %q not found", hookCfg.Name)
		}

		// Try to create a new hook
		hook, err := hookFactory.New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create hook %q: %w", hookCfg.Name, err)
		}

		// Set up the hook
		err = hook.Setup(ctx, hookCfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup hook %q: %w", hookCfg.Name, err)
		}
		slog.Debug("Hook set up", "name", hookCfg.Name)

		// Save the teardown function
		teardowns[hookCfg.Name] = hook.Teardown
	}
	return teardownFunc, nil
}
