package config

type (
	Path struct {
		Methods      []Method `yaml:"methods"`
		Middlewares  Plugins  `yaml:"middlewares"`
		Redirect     Redirect `yaml:"redirect"`
		ReqModifiers Plugins  `yaml:"req-modifiers"`
	}
	Paths map[string]Path

	Redirect struct {
		Backends []Backend `yaml:"backends"`
	}
)
