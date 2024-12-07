package config

type (
	Path struct {
		Methods      []Method `yaml:"methods"`
		Middlewares  Plugins  `yaml:"middlewares"`
		ReqModifiers Plugins  `yaml:"req-modifiers"`
	}
	Paths map[string]Path
)
