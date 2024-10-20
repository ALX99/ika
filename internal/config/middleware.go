package config

import "iter"

type (
	Middleware struct {
		Name   string         `yaml:"name"`
		Config map[string]any `yaml:"config"`
	}
	Middlewares []Middleware
)

// Names returns an iterator that yields the names of all middlewares.
func (m Middlewares) Names() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, mw := range m {
			if !yield(mw.Name) {
				return
			}
		}
	}
}
