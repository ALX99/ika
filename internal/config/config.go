package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Servers    []Server   `yaml:"servers"`
	Namespaces Namespaces `yaml:"namespaces"`
	Ika        Ika        `yaml:"ika"`
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

	return cfg, nil
}
