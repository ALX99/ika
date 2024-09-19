package plugin

import (
	"context"
	"net/http"
)

//go:generate minimock -i Factory,TransportHook,Hook -o ../mocks -s _mock.go

type Factory[T any] interface {
	New(context.Context) (T, error)
}

type Hook interface {
	Setup(ctx context.Context, config map[string]any) error
	Teardown(ctx context.Context) error
}

type MiddlewareHook interface {
	HookMiddleware(ctx context.Context, name string, next http.Handler) (http.Handler, error)
}

type FirstHandlerHook interface {
	HookFirstHandler(ctx context.Context, handler http.Handler) (http.Handler, error)
}

type TransportHook interface {
	HookTransport(ctx context.Context, transport http.RoundTripper) (http.RoundTripper, error)
}
