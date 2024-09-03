package config

import "iter"

type (
	Path struct {
		RewritePath Nullable[string] `yaml:"rewritePath"`
		Methods     []Method         `yaml:"methods"`
		Backends    []Backend        `yaml:"backends"`
		Middlewares Middlewares      `yaml:"middlewares"`
	}
	Paths map[string]Path
)

// MergedMiddlewares returns an iterator that yields all enabled middlewares from the path and namespace.
func (p Path) MergedMiddlewares(ns Namespace) iter.Seq[Middleware] {
	return func(yield func(Middleware) bool) {
		for mw := range p.Middlewares.enabled() {
			if !yield(mw) {
				return
			}
		}

		for mw := range ns.Middlewares.enabled() {
			if !yield(mw) {
				return
			}
		}
	}
}
