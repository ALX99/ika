package plugin

import (
	"context"
	"log/slog"
	"net/http"
)

// TODO
type MiddlewareHook interface {
	// HookMiddleware wraps the provided HTTP handler with custom middleware logic.
	HookMiddleware(ctx context.Context, name string, next http.Handler) (http.Handler, error)
}

// InjectionLevel defines the granularity of plugin injection.
type (
	InjectionLevel uint8
)

const (
	// LevelPath specifies injection at a specific path level.
	LevelPath InjectionLevel = iota

	// LevelNamespace specifies injection at a namespace level.
	LevelNamespace
)

// Factory is responsible for creating plugin instances.
type Factory interface {
	// Name returns the name of the plugin created by this factory.
	Name() string

	// New creates and returns a new instance of the plugin.
	// Each call to New can return either a new or shared instance, but using shared instances may cause debugging difficulties.
	New(ctx context.Context, ictx InjectionContext) (Plugin, error)
}

// InjectionContext provides details about the context in which a plugin is injected.
type InjectionContext struct {
	// Namespace indicates the namespace where the plugin is injected.
	// Empty if not injected at the namespace or path level.
	Namespace string

	// PathPattern specifies the path pattern where the plugin is injected.
	// Empty if not injected at the path level.
	PathPattern string

	// Level indicates whether the injection is at the namespace or path level.
	Level InjectionLevel

	// Logger is the logger meant for the plugin.
	Logger *slog.Logger
}

// Plugin defines the common interface for all plugins in Ika.
type Plugin interface {
	// Setup initializes the plugin with the given configuration and context.
	// If injected multiple times at the same level, Setup will be called multiple times.
	//
	// If injected at a level where the plugin can not operate, an error should be returned.
	Setup(ctx context.Context, ictx InjectionContext, config map[string]any) error

	// Teardown cleans up resources when the plugin is no longer needed.
	Teardown(ctx context.Context) error
}

// RequestModifier allows plugins to modify incoming HTTP requests before processing.
type RequestModifier interface {
	Plugin

	// ModifyRequest processes and returns the modified HTTP request.
	ModifyRequest(r *http.Request) (*http.Request, error)
}

// Middleware enables plugins to modify both requests and responses.
type Middleware interface {
	Plugin

	// Handler wraps the given handler with custom logic for processing requests and responses.
	Handler(next Handler) Handler
}

// TransportHooker allows plugins to modify the transport mechanism used by Ika.
type TransportHooker interface {
	Plugin

	// HookTransport returns a new or modified HTTP transport.
	// It can wrap or replace the existing transport.
	HookTransport(roundtripper http.RoundTripper) http.RoundTripper
}

// FirstHandlerHooker enables plugins to hijack the first handler executed for a request.
//
// It is semantically equivalent to a middleware, but is executed before all other middleware
// and thus is useful for things such as tracing or logging.
//
// If multiple FirstHandlerHookers are injected, an error will be returned.
type FirstHandlerHooker interface {
	Plugin
	Middleware
}
