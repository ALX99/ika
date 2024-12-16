package config

import (
	"time"
)

type Ika struct {
	Logger                  Logger        `yaml:"logger"`
	GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
}
