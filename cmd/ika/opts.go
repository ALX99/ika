//go:build full

package main

import (
	"github.com/alx99/ika/gateway"
	"github.com/alx99/ika/plugins"
)

func init() {
	opts = append(opts, gateway.WithPlugin(plugins.ReqModifier{}))
	opts = append(opts, gateway.WithPlugin(plugins.AccessLogger{}))
}
