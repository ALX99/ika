//go:build full

package main

import (
	"github.com/alx99/ika/cmd/option"
	"github.com/alx99/ika/plugins"
)

func init() {
	opts = append(opts, option.WithPlugin(plugins.ReqModifier{}))
	opts = append(opts, option.WithPlugin(plugins.AccessLogger{}))
}
