package ika

import "github.com/alx99/ika/hook"

type startCfg struct {
	hooks map[string]hook.Hooker
}

// Option represents an option for Run.
type Option func(*startCfg)

func WithHook(name string, hook hook.Hooker) Option {
	return func(cfg *startCfg) {
		cfg.hooks[name] = hook
	}
}
