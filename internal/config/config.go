package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server            Server     `yaml:"server"`
	Namespaces        Namespaces `yaml:"namespaces"`
	NamespaceOverride Namespace  `yaml:"namespaceOverride"`
	Ika               Ika        `yaml:"ika"`
}

func Read(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	return cfg, yaml.NewDecoder(f).Decode(&cfg)
}

// ApplyOverride applies the NamespaceOverrides to the Config.
func (c *Config) ApplyOverride() {
	for name, ns := range c.Namespaces {
		ns.Transport = override(ns.Transport, c.NamespaceOverride.Transport)
		c.Namespaces[name] = ns
	}
}
