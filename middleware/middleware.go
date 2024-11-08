package middleware

import (
	"context"
	"net/http"

	"github.com/alx99/ika/internal/middleware"
)

type MiddlewareFunc func(http.Handler) http.Handler

// RegisterProvider registers a provider which provides a middleware with the specified name.
func RegisterProvider(name string, provider middleware.Provider) error {
	return middleware.Register(name, provider)
}

// RegisterFunc creates a [MiddlewareFunc] the given middleware func.
func RegisterFunc(name string, fun MiddlewareFunc) error {
	return RegisterProvider(name, basicProvider{middleware: fun})
}

type basicProvider struct {
	middleware MiddlewareFunc
}

func (p basicProvider) New(_ context.Context, next http.Handler) (middleware.Middleware, error) {
	return basicMiddleware(func(w http.ResponseWriter, r *http.Request) {
		p.middleware(next).ServeHTTP(w, r)
	}), nil
}

type basicMiddleware http.HandlerFunc

func (basicMiddleware) Setup(_ context.Context, _ map[string]any) error     { return nil }
func (basicMiddleware) Teardown(_ context.Context) error                    { return nil }
func (mw basicMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) { mw(w, r) }
