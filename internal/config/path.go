package config

type (
	Path struct {
		Methods      []Method `yaml:"methods" json:"methods"`
		Middlewares  Plugins  `yaml:"middlewares" json:"middlewares"`
		ReqModifiers Plugins  `yaml:"reqModifiers" json:"reqModifiers"`
	}
	Paths map[string]Path
)
