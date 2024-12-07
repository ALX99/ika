package config

type (
	Path struct {
		Methods      []Method `yaml:"methods"`
		Middlewares  Plugins  `yaml:"middlewares"`
		ReqModifiers Plugins  `yaml:"reqModifiers"`
	}
	Paths map[string]Path
)
