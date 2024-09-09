package middleware

import (
	"context"
	"net/http"
)

type (
	Middleware interface {
		// Setup should initialize the middleware based on the given configuration.
		Setup(ctx context.Context, next http.Handler, config map[string]any) error
		// Teardown should clean up any potential resources used by the middleware.
		Teardown(ctx context.Context) error
		http.Handler
	}
	MiddlewareFunc func(http.Handler) http.Handler
)
