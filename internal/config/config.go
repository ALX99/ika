package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Ika struct {
	Server                  Server        `yaml:"server"`
	Namespaces              Namespaces    `yaml:"namespaces"`
	GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
}

func Read(path string) (Ika, error) {
	cfg := Ika{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	return cfg, yaml.NewDecoder(f).Decode(&cfg)
}
