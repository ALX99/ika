package config

type (
	Path struct {
		Methods     []Method    `yaml:"methods"`
		Middlewares Middlewares `yaml:"middlewares"`
		Redirect    Redirect    `yaml:"redirect"`
		Plugins     Plugins     `yaml:"plugins"`
	}
	Paths map[string]Path

	Redirect struct {
		Backends []Backend `yaml:"backends"`
	}
)
