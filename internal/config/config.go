package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     Server     `yaml:"server"`
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

	return cfg, yaml.NewDecoder(f).Decode(&cfg)
}
