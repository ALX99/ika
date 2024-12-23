package request

import (
	"context"
	"net/http"
)

type keyPathTrim struct{}

func GetPath(r *http.Request) string {
	path := r.URL.RawPath
	if path == "" {
		path = r.URL.Path
	}
	return path
}

func GetPathToTrim(r *http.Request) string {
	return r.Context().Value(keyPathTrim{}).(string)
}

func SetPathToTrim(r *http.Request, route string) {
	*r = *r.WithContext(context.WithValue(r.Context(), keyPathTrim{}, route))
}
