package config

import "iter"

type (
	Middleware struct {
		Name    string            `yaml:"name"`
		Enabled Defaultable[bool] `yaml:"enabled"`
		Args    map[string]any    `yaml:",inline"`
	}
	Middlewares []Middleware
)

// Names returns an iterator that yields the names of all enabled middlewares.
func (m Middlewares) Names() iter.Seq[string] {
	return func(yield func(string) bool) {
		for mw := range m.enabled() {
			if !yield(mw.Name) {
				return
			}
		}
	}
}

// enabled returns an iterator that yields all enabled middlewares.
func (m Middlewares) enabled() iter.Seq[Middleware] {
	return func(yield func(Middleware) bool) {
		for _, mw := range m {
			if mw.Enabled.Or(true) {
				if !yield(mw) {
					return
				}
			}
		}
	}
}
