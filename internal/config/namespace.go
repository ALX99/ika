package config

type (
	Namespace struct {
		Transport    Transport `yaml:"transport"`
		Paths        Paths     `yaml:"paths"`
		Middlewares  Plugins   `yaml:"middlewares"` // todo add support
		ReqModifiers Plugins   `yaml:"req-modifiers"`
		Plugins      Plugins   `yaml:"plugins"`
	}
	Namespaces map[string]Namespace
)
