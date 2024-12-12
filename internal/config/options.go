package config

import "github.com/alx99/ika/plugin"

type Options struct {
	Plugins map[string]plugin.Factory
}
