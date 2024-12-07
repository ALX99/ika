package config

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"

	"github.com/alx99/ika/internal/plugins"
	pplugin "github.com/alx99/ika/plugin"
	"gopkg.in/yaml.v3"
)

type PluginFactory struct {
	PluginVal reflect.Value

	// name of the plugin
	name string
	// namespaces is a list of namespaces where the plugin is enabled
	namespaces []string
}

type RunOpts struct {
	Plugins2 []pplugin.Factory
}

type Config struct {
	Servers           []Server   `yaml:"servers"`
	Namespaces        Namespaces `yaml:"namespaces"`
	NamespaceOverride Namespace  `yaml:"namespaceOverrides"`
	Ika               Ika        `yaml:"ika"`

	// runtime configuration
	pluginFactories []PluginFactory

	PluginFacs2 map[string]pplugin.Factory
}

func Read(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}

	defer f.Close()
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return cfg, err
	}

	if len(cfg.Servers) < 1 {
		return cfg, errors.New("at least one server must be specified")
	}

	cfg.ApplyOverride()
	return cfg, nil
}

func NewRunOpts() RunOpts {
	return RunOpts{}
}

func (c *Config) SetRuntimeOpts(opts RunOpts) error {
	if err := c.loadPlugins(opts.Plugins2); err != nil {
		return err
	}

	return nil
}

func (c *Config) loadPlugins(factories2 []pplugin.Factory) error {
	c.PluginFacs2 = make(map[string]pplugin.Factory)
	c.PluginFacs2["basic-modifier"] = plugins.ReqModifier{} // hack
	c.PluginFacs2["accessLog"] = plugins.AccessLogger{}     // hack
	for _, factory := range factories2 {
		c.PluginFacs2[factory.Name()] = factory
	}

	for _, ns := range c.Namespaces {
		for _, path := range ns.Paths {
			for plugin := range path.ReqModifiers.Enabled() {
				if !slices.Contains(slices.Collect(maps.Keys(c.PluginFacs2)), plugin.Name) {
					return fmt.Errorf("plugin %q not found", plugin.Name)
				}
			}
		}
	}

	return nil
}

// ApplyOverride applies the NamespaceOverrides to the Config.
func (c *Config) ApplyOverride() {
	for name, ns := range c.Namespaces {
		ns.Transport = override(ns.Transport, c.NamespaceOverride.Transport)
		c.Namespaces[name] = ns
	}
}
