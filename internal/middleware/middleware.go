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

type Middleware interface {
	// Setup should initialize the middleware based on the given configuration.
	Setup(ctx context.Context, config map[string]any) error
	// Teardown should clean up any potential resources used by the middleware.
	Teardown(ctx context.Context) error
	http.Handler
}

// Provider is able to create new middlewares
type Provider interface {
	// New should return a new middleware
	New(ctx context.Context, next http.Handler) (Middleware, error)
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

// Get creates and returns a middleware provider by its name.
func Get(ctx context.Context, name string, next http.Handler) (Middleware, error) {
	mu.RLock()
	defer mu.RUnlock()
	m, ok := providers[name]
	if !ok {
		return nil, fmt.Errorf("middleware %q is not registered", name)
	}
	mw, err := m.New(ctx, next)
	if err != nil {
		return nil, fmt.Errorf("could not create middleware %q: %w", name, err)
	}
	return mw, nil
}
