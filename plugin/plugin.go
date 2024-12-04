package plugin

import (
	"context"
	"net/http"
)

//go:generate minimock -i Factory,TransportHook,Hook -o ../mocks -s _mock.go

type Factory interface {
	New(context.Context) (any, error)
}

type Setupper interface {
	Setup(ctx context.Context, config map[string]any) error
}

type Teardowner interface {
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

type (
	Capability     uint16
	InjectionLevel uint8
)

const (
	// The plugins which report this capability must implement the [RequestModifier] interface
	CapModifyRequests Capability = iota

	PathLevel InjectionLevel = iota
)

type NFactory interface {
	New(ctx context.Context) (Plugin, error)
}

// InjectionContext contains information about the context the plugin was injected into.
type InjectionContext struct {
	// The namespace the plugin is injected into
	// If it was not injected on a namespace or path level, it will be empty.
	Namespace string
	// The path pattern the plugin as injected into.
	// If it was not injected on a path level, it will be empty.
	PathPattern string
	// The level of where the plugin was injected.
	Level InjectionLevel
}

type Plugin interface {
	// Name must return the name of the plugin
	Name() string
	// Capabilities must return the capabilities of the plugin.
	Capabilities() []Capability
	// Setup should do the necessary setup for the plugin given the configuration.
	// In case the plugin is injected multiple times, this function will be called for each injection.
	// It is up to the plugin itself, to handle this correctly.
	Setup(ctx context.Context, context InjectionContext, config map[string]any) error
	// Teardown should do the necessary teardown for the plugin.
	Teardown(ctx context.Context) error
}

type RequestModifier interface {
	ModifyRequest(ctx context.Context, r *http.Request) (*http.Request, error)
}
