package config

type (
	Namespace struct {
		Transport    Transport `yaml:"transport" json:"transport"`
		NSPaths      []string  `yaml:"nsPaths" json:"nsPaths"`
		Paths        Paths     `yaml:"paths" json:"paths"`
		Middlewares  Plugins   `yaml:"middlewares" json:"middlewares"`
		ReqModifiers Plugins   `yaml:"reqModifiers" json:"reqModifiers"`
		Hooks        Plugins   `yaml:"hooks" json:"hooks"`
	}
	Namespaces map[string]Namespace
)
