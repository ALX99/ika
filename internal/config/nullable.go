package config

import "gopkg.in/yaml.v3"

type Nullable[T any] struct {
	V   T
	set bool
}

func (n *Nullable[T]) UnmarshalYAML(value *yaml.Node) error {
	type alias Nullable[T]
	tmp := alias{}

	if err := value.Decode(&tmp.V); err != nil {
		return err
	}

	n.set = true
	*n = Nullable[T](tmp)
	return nil
}

// Or returns the value of the underlying type if it is set, otherwise it returns the default value.
func (n Nullable[T]) Or(defaultV T) T {
	if !n.set {
		return defaultV
	}
	return n.V
}

func NewNullable[T any](v T) Nullable[T] {
	return Nullable[T]{V: v, set: true}
}
