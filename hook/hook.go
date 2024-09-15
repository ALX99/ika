package hook

import (
	"context"
	"net/http"
)

type Factory interface {
	New(context.Context) (Hook, error)
}

type Hook interface {
	Setup(ctx context.Context, config map[string]any) error
	Teardown(ctx context.Context) error
}

type TransportHook interface {
	HookTransport(ctx context.Context, transport http.RoundTripper) (http.RoundTripper, error)
}
