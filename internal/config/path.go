package config

type (
	Route struct {
		Methods      []Method `json:"methods"`
		Middlewares  Plugins  `json:"middlewares"`
		ReqModifiers Plugins  `json:"reqModifiers"`
	}
	Routes map[string]Route
)
