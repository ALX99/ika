package middleware

import (
	"context"
	"net/http"
)

type (
	metaDataKey struct{}
	Metadata    struct {
		Namespace      string
		Route          string
		GeneratedRoute string
	}
)

// BindMetadata binds some metadata to the request context.
func BindMetadata(data Metadata, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), metaDataKey{}, data)))
	})
}

func GetMetadata(ctx context.Context) Metadata {
	data, ok := ctx.Value(metaDataKey{}).(Metadata)
	if !ok {
		return Metadata{}
	}
	return data
}
