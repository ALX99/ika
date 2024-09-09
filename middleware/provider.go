package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

var (
	providers = make(map[string]Provider)
	mu        = sync.RWMutex{}
)

type Provider interface {
	// NewMiddleware should return a new middleware
	NewMiddleware(ctx context.Context) (Middleware, error)
}

// Register registers a new provider which provides a middleware with the specified name.
func Register(name string, provider Provider) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := providers[name]; ok {
		return fmt.Errorf("middleware %q is already registered", name)
	}
	providers[name] = provider
	return nil
}

// Get returns a middleware provider by its name.
func Get(name string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := providers[name]
	return m, ok
}

// FromMiddlewareFunc creates a new provider from the given middleware func.
func FromMiddlewareFunc(fun MiddlewareFunc) Provider {
	return basicProvider{middleware: fun}
}

type basicProvider struct {
	middleware func(http.Handler) http.Handler
}

func (p basicProvider) NewMiddleware(_ context.Context) (Middleware, error) {
	return &basicMiddleware{middleware: p.middleware}, nil
}

type basicMiddleware struct {
	middleware func(http.Handler) http.Handler
	handler    http.Handler
}

func (m *basicMiddleware) Setup(_ context.Context, next http.Handler, _ map[string]any) error {
	m.handler = m.middleware(next)
	return nil
}

func (m *basicMiddleware) Teardown(_ context.Context) error {
	return nil
}

func (m *basicMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.handler.ServeHTTP(w, r)
}
