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

type Middleware func(http.Handler) http.Handler

type Provider interface {
	// GetMiddleware should set up the middleware based on the given configuration
	GetMiddleware(ctx context.Context, config map[string]any) (Middleware, error)
	// Teardown should clean up any potential resources used by the middleware.
	// It is called when the server is shutting down.
	Teardown(ctx context.Context) error
}

// RegisterProvider registers a new provider which provides a middleware with the specified name.
func RegisterProvider(name string, provider Provider) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := providers[name]; ok {
		return fmt.Errorf("middleware %q is already registered", name)
	}
	providers[name] = provider
	return nil
}

// Get returns a middleware by name.
func Get(name string) (Provider, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := providers[name]
	return m, ok
}

// Len returns the number of registered middlewares.
func Len() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(providers)
}
