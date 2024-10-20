package config

type (
	Path struct {
		RewritePath Nullable[string] `yaml:"rewritePath"`
		Methods     []Method         `yaml:"methods"`
		Backends    []Backend        `yaml:"backends"`
		Middlewares Middlewares      `yaml:"middlewares"`
	}
	Paths map[string]Path
)
