package config

import "github.com/alx99/ika"

type Options struct {
	Plugins  map[string]ika.PluginFactory
	Validate bool
}
