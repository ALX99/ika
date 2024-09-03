package config

import "time"

type Ika struct {
	GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
}
