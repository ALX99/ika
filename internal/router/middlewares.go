package router

import (
	"context"
	"net/http"
)

type nsKey struct{}

func bindNamespace(ns string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), nsKey{}, ns)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetNamespace returns the namespace from the given context.
// In case the namespace is not set, it returns an empty string.
func GetNamespace(ctx context.Context) string {
	ns, _ := ctx.Value(nsKey{}).(string)
	return ns
}
