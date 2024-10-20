package config

type (
	Path struct {
		Methods     []Method    `yaml:"methods"`
		Middlewares Middlewares `yaml:"middlewares"`
		Redirect    Redirect    `yaml:"redirect"`
	}
	Paths map[string]Path

	Redirect struct {
		Paths    []string  `yaml:"paths"`
		Backends []Backend `yaml:"backends"`
	}
)
