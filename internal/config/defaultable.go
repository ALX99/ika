package config

import "gopkg.in/yaml.v3"

type nullable[T any] struct {
	v   T
	set bool
}

func (n *nullable[T]) UnmarshalYAML(value *yaml.Node) error {
	type alias nullable[T]
	tmp := alias{}

	if err := value.Decode(&tmp.v); err != nil {
		return err
	}

	n.set = true
	*n = nullable[T](tmp)
	return nil
}

type Defaultable[T any] struct {
	v nullable[T]
}

func (d *Defaultable[T]) UnmarshalYAML(value *yaml.Node) error {
	return d.v.UnmarshalYAML(value)
}

// Or returns the value of the underlying type if it is set, otherwise it returns the default value.
func (d *Defaultable[T]) Or(defaultV T) T {
	if !d.v.set {
		return defaultV
	}
	return d.v.v
}
