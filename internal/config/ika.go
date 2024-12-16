package config

type Ika struct {
	Logger                  Logger   `yaml:"logger" json:"logger"`
	GracefulShutdownTimeout Duration `yaml:"gracefulShutdownTimeout" json:"gracefulShutdownTimeout"`
}
