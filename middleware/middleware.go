package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

var (
	middlewares = make(map[string]Middleware)
	mu          = sync.RWMutex{}
)

type Middleware interface {
	Setup(ctx context.Context, config map[string]any) error
	Teardown(ctx context.Context) error
	Handle(http.Handler) http.Handler
}

// Register registers a middleware with the given name.
func Register(name string, middleware Middleware) error {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := middlewares[name]; ok {
		return fmt.Errorf("middleware %q is already registered", name)
	}
	middlewares[name] = middleware
	return nil
}

// Get returns a middleware by name.
func Get(name string) (Middleware, bool) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := middlewares[name]
	return m, ok
}

// Len returns the number of registered middlewares.
func Len() int {
	mu.RLock()
	defer mu.RUnlock()
	return len(middlewares)
}
