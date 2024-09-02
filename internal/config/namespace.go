package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type (
	Namespace struct {
		Name                   string      `yaml:"name"`
		Backends               []Backend   `yaml:"backends"`
		Transport              Transport   `yaml:"transport"`
		Paths                  Paths       `yaml:"paths"`
		Middlewares            Middlewares `yaml:"middlewares"`
		Hosts                  []string    `yaml:"hosts"`
		DisableNamespacedPaths bool        `yaml:"disableNamespacedPaths"`
	}
	Namespaces map[string]Namespace
)

func (ns *Namespace) UnmarshalYAML(value *yaml.Node) error {
	type alias Namespace
	tmp := alias{}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	if len(tmp.Hosts) == 0 && tmp.DisableNamespacedPaths {
		return fmt.Errorf("namespace has no hosts and namespaced paths are disabled")
	}

	*ns = Namespace(tmp)
	return nil
}

func (ns *Namespaces) UnmarshalYAML(value *yaml.Node) error {
	tmp := make(map[string]Namespace)

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	for name, ns := range tmp {
		ns.Name = name
		tmp[name] = ns
	}

	*ns = tmp
	return nil
}
