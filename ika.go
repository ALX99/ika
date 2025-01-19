package ika

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alx99/ika/internal/http/request"
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

// PluginFactory is responsible for creating plugin instances.
type PluginFactory interface {
	// Name returns the name of the plugin created by this factory.
	Name() string

	// New creates and returns a new instance of the plugin.
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
	//
	// If injected multiple times at the same level, Setup will be called multiple times.
	// If injected at a level where the plugin can not operate, an error should be returned.
	Setup(ctx context.Context, ictx InjectionContext, config map[string]any) error

	// Teardown cleans up potential resources used by the plugin.
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

// TripperHooker allows plugins to modify the [http.RoundTripper] used by Ika.
type TripperHooker interface {
	Plugin

	// HookTripper returns a new or modified [http.RoundTripper].
	// It can wrap or replace the existing transport.
	HookTripper(tripper http.RoundTripper) (http.RoundTripper, error)
}

// FirstHandlerHooker enables plugins to hijack the first handler executed for a request.
//
// It is semantically equivalent to a middleware, but is executed before all other middleware
// and thus is useful for things such as tracing or logging.
type FirstHandlerHooker interface {
	Plugin
	Middleware
}

// Handler is identical to [http.Handler] except that it is able to return an error.
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as [Handler]s.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// FromHTTPHandler turns an [http.Handler] into an [Handler].
func FromHTTPHandler(h http.Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		h.ServeHTTP(w, r)
		return nil
	})
}

// ToHTTPHandler converts an [Handler] into an [http.Handler] using [HandlerFunc.ToHTTPHandler].
func ToHTTPHandler(h Handler, errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.ServeHTTP(w, r)
		if err != nil {
			if errorHandler != nil {
				errorHandler(w, r, err)
				return
			}
			defualtErrorHandler(w, r, err)
		}
	})
}

// ToHTTPHandler converts an [Handler] into an [http.Handler].
// If the function returns an error, it will be written to the response using the provided error handler.
// If the error handler is nil, the error will be written as a 500 Internal Server Error.
func (f HandlerFunc) ToHTTPHandler(errorHandler func(http.ResponseWriter, *http.Request, error)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f.ServeHTTP(w, r)
		if err != nil {
			if errorHandler != nil {
				errorHandler(w, r, err)
				return
			}
			defualtErrorHandler(w, r, err)
		}
	})
}

func defualtErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	slog.LogAttrs(r.Context(),
		slog.LevelError,
		"Error handling request",
		slog.String("path", request.GetPath(r)),
		slog.String("error", err.Error()))
	http.Error(w, "failed to handle request", http.StatusInternalServerError)
}
