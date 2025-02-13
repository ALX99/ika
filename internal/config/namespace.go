package config

type (
	Namespace struct {
		Transport    Transport `json:"transport"`
		Mounts       []string  `json:"mounts"`
		Routes       Routes    `json:"routes"`
		Middlewares  Plugins   `json:"middlewares"`
		ReqModifiers Plugins   `json:"reqModifiers"`
		Hooks        Plugins   `json:"hooks"`
	}
	Namespaces map[string]Namespace
)
