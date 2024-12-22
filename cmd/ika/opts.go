//go:build full

package main

import (
	"github.com/alx99/ika"
	"github.com/alx99/ika/plugins"
)

func init() {
	opts = append(opts, ika.WithPlugin(plugins.ReqModifier{}))
	opts = append(opts, ika.WithPlugin(plugins.AccessLogger{}))
}
