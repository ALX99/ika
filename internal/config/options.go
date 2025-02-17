package config

import "github.com/alx99/ika"

type ComptimeOpts struct {
	Plugins  map[string]ika.PluginFactory
	Validate bool
}
