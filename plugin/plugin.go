package plugin

import (
	"context"
	"net/http"
)

// TODO
type MiddlewareHook interface {
	HookMiddleware(ctx context.Context, name string, next http.Handler) (http.Handler, error)
}

type (
	InjectionLevel uint8
)

const (
	LevelPath InjectionLevel = iota
	LevelNamespace
)

type Factory interface {
	// Name must return the name of the plugin that the factory creates.
	Name() string

	// New returns an instance of the plugin.
	//
	// It is allowed to return the same instance for multiple calls.
	// However, do note that this might lead to difficult to debug issues.
	// For this reason, it is recommended to return a new instance for each call.
	New(ctx context.Context, iCtx InjectionContext) (Plugin, error)
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

// RequestModifier is an interface that plugins can implement to modify requests.
type RequestModifier interface {
	Plugin
	ModifyRequest(r *http.Request) (*http.Request, error)
}

// Middleware is an interface that plugins can implement to modify both requests and responses.
type Middleware interface {
	Plugin
	Handler(next ErrHandler) ErrHandler
}

// TransportHooker is an interface that plugins can implement to modify the transport that ika uses.
type TransportHooker interface {
	Plugin
	HookTransport(transport http.RoundTripper) http.RoundTripper
}

// FirstHandlerHooker is an interface that plugins can implement to modify the first handler that get executed for each path.
type FirstHandlerHooker interface {
	Plugin
	HookFirstHandler(next ErrHandler) ErrHandler
}
