package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type RunOpts struct {
	Hooks map[string]any
}

type Config struct {
	Server            Server     `yaml:"server"`
	Namespaces        Namespaces `yaml:"namespaces"`
	NamespaceOverride Namespace  `yaml:"namespaceOverrides"`
	Ika               Ika        `yaml:"ika"`
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

	cfg.ApplyOverride()
	return cfg, nil
}

func NewRunOpts() RunOpts {
	return RunOpts{
		Hooks: make(map[string]any),
	}
}

// ApplyOverride applies the NamespaceOverrides to the Config.
func (c *Config) ApplyOverride() {
	for name, ns := range c.Namespaces {
		ns.Transport = override(ns.Transport, c.NamespaceOverride.Transport)
		c.Namespaces[name] = ns
	}
}
