package config

type (
	Namespace struct {
		Transport    Transport `json:"transport"`
		Mounts       []string  `json:"mounts"`
		Paths        Paths     `json:"paths"`
		Middlewares  Plugins   `json:"middlewares"`
		ReqModifiers Plugins   `json:"reqModifiers"`
		Hooks        Plugins   `json:"hooks"`
	}
	Namespaces map[string]Namespace
)
