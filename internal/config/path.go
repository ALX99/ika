package config

type (
	Path struct {
		Methods      []Method `json:"methods"`
		Middlewares  Plugins  `json:"middlewares"`
		ReqModifiers Plugins  `json:"reqModifiers"`
	}
	Paths map[string]Path
)
