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
	// The plugins which report this capability must implement the [Middleware] interface
	CapMiddleware

	LevelPath InjectionLevel = iota
)

type NFactory interface {
	// New returns an instance of the plugin.
	//
	// It is allowed to return the same instance for multiple calls.
	// However, do note that this might lead to difficult to debug issues.
	// For this reason, it is recommended to return a new instance for each call.
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
	//
	// It is used for validation purposes to ensure the plugin satisfies
	// the correct interface(s).
	Capabilities() []Capability

	// InjectionLevels must return the levels of injection the plugin supports.
	InjectionLevels() []InjectionLevel

	// Setup should do the necessary setup for the plugin given the configuration.
	//
	// In case the same plugin is injected on multiple multiple levels [Plugin.Setup] will be called multiple times.
	// It is up to individual plugins to handle this case.
	Setup(ctx context.Context, iCtx InjectionContext, config map[string]any) error

	// Teardown should do the necessary teardown for the plugin.
	// It is called once the plugin is no longer needed.
	Teardown(ctx context.Context) error
}

// RequestModifier is an interface that plugins can implement to modify incoming requests.
type RequestModifier interface {
	Plugin
	ModifyRequest(ctx context.Context, r *http.Request) (*http.Request, error)
}

// Middleware is an interface that plugins can implement to modify both requests and responses.
type Middleware interface {
	Plugin
	// Handler should return an [ErrHandler] which will be called for each request.
	Handler(next ErrHandler) (ErrHandler, error)
}
