package config

type Ika struct {
	Logger                  Logger   `json:"logger"`
	GracefulShutdownTimeout Duration `json:"gracefulShutdownTimeout"`
}
