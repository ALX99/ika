package config

import (
	"time"

	"github.com/alx99/ika/internal/logger"
)

type Ika struct {
	Logger                  logger.Config `yaml:"logger"`
	GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
}
