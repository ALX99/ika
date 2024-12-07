package config

type (
	Namespace struct {
		Transport    Transport `yaml:"transport"`
		Paths        Paths     `yaml:"paths"`
		Middlewares  Plugins   `yaml:"middlewares"` // todo add support
		ReqModifiers Plugins   `yaml:"reqModifiers"`
		Hooks        Plugins   `yaml:"hooks"`
	}
	Namespaces map[string]Namespace
)
