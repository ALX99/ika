package config

type (
	Namespace struct {
		Backends               []Backend      `yaml:"backends"`
		Transport              Transport      `yaml:"transport"`
		Paths                  Paths          `yaml:"paths"`
		Middlewares            Middlewares    `yaml:"middlewares"`
		Plugins                Plugins        `yaml:"plugins"`
		Hosts                  []string       `yaml:"hosts"`
		DisableNamespacedPaths Nullable[bool] `yaml:"disableNamespacedPaths"`
	}
	Namespaces map[string]Namespace
)
