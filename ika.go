package ika

import (
	"cmp"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/http/request"
)

const (
	// ScopeRoute indicates that the plugin is injected at the route scope.
	ScopeRoute InjectionLevel = iota

	// ScopeNamespace indicates that the plugin is injected at the namespace scope.
	ScopeNamespace
)

// InjectionLevel represents the granularity at which a plugin is injected.
// It determines whether a plugin is applied at a route or namespace scope.
type InjectionLevel uint8

// ErrorHandler is a function that handles errors that occur during request processing.
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// PluginFactory creates new instances of a plugin.
type PluginFactory interface {
	// Name returns the unique name of the plugin produced by this factory.
	Name() string

	// New instantiates a new plugin using the provided injection context and configuration.
	New(ctx context.Context, ictx InjectionContext, config map[string]any) (Plugin, error)
}

// InjectionContext contains information about the context in which a plugin is injected.
type InjectionContext struct {
	// Namespace specifies the target namespace for plugin injection.
	// It is empty when the plugin is not injected on a namespace or route level.
	Namespace string

	// Route specifies the target route for plugin injection.
	// It is empty when the plugin is not injected on a route level.
	Route string

	// TODO: provide mux pattern

	// Scope indicates the injection level at which the plugin is applied.
	// It can be either ScopeRoute or ScopeNamespace.
	Scope InjectionLevel

	// Logger is the logger allocated for the plugin.
	Logger *slog.Logger
}

// Plugin is the common interface for all plugins in Ika.
type Plugin interface {
	// Teardown releases any resources allocated by the plugin.
	Teardown(ctx context.Context) error
}

// RequestModifier allows plugins to modify incoming HTTP requests before processing.
type RequestModifier interface {
	Plugin

	// ModifyRequest allows the plugin to modify the incoming HTTP request prior to processing.
	ModifyRequest(r *http.Request) error
}

// Middleware enables plugins to modify both requests and responses.
type Middleware interface {
	Plugin

	// Handler wraps the given HTTP handler with additional plugin-specific logic.
	Handler(next Handler) Handler
}

// TripperHook allows plugins to modify or wrap the http.RoundTripper used by Ika.
type TripperHook interface {
	Plugin

	// HookTripper returns a new or modified http.RoundTripper.
	// It may wrap or outright replace the provided transport.
	HookTripper(rt http.RoundTripper) (http.RoundTripper, error)
}

// OnRequestHook enables plugins to execute hooks immediately when a request is received.
// It is functionally equivalent to Middleware but is invoked before other middleware,
// making it ideal for tasks such as tracing or logging.
type OnRequestHook interface {
	Plugin
	Middleware
}

// Handler is similar to http.Handler but its ServeHTTP method may return an error.
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is an adapter to allow the use of ordinary functions as Handlers.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return f(w, r)
}

// ToHTTPHandler converts a Handler into a standard http.Handler.
// If the Handler returns an error during execution, the provided error handler is invoked.
// If errHandler is nil, DefaultErrorHandler is used.
func ToHTTPHandler(h Handler, errHandler ErrorHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.ServeHTTP(w, r)
		if err != nil {
			if errHandler != nil {
				errHandler(w, r, err)
				return
			}
			DefaultErrorHandler(w, r, err)
		}
	})
}

// ToHTTPHandler converts a HandlerFunc into a standard http.Handler.
// If the function returns an error, the error is written to the response using the provided error handler,
// or DefaultErrorHandler if errHandler is nil.
func (f HandlerFunc) ToHTTPHandler(errHandler ErrorHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f.ServeHTTP(w, r)
		if err != nil {
			if errHandler != nil {
				errHandler(w, r, err)
				return
			}
			DefaultErrorHandler(w, r, err)
		}
	})
}

// DefaultErrorHandler writes an error response to the client, formatting the response based on the client's Accept header.
// If the client accepts "application/json" or "*/*", the error is encoded in JSON; otherwise, plain text is used.
// If the error implements Status(), TypeURI(), Title(), and Detail() methods, their values will be used in the response.
func DefaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	logWriteError := func(err error) {
		slog.LogAttrs(r.Context(),
			slog.LevelError,
			"Error writing error response",
			slog.String("error", err.Error()))
	}

	errorResp := struct {
		Type   string `json:"type,omitzero"`
		Title  string `json:"title,omitzero"`
		Detail string `json:"detail,omitzero"`
		Status int    `json:"status,omitzero"`
	}{}

	errorResp.Status = http.StatusInternalServerError

	if err, ok := err.(interface{ Status() int }); ok {
		errorResp.Status = err.Status()
	}

	if err, ok := err.(interface{ TypeURI() string }); ok {
		errorResp.Type = err.TypeURI()
	}

	if err, ok := err.(interface{ Title() string }); ok {
		errorResp.Title = err.Title()
	}

	if err, ok := err.(interface{ Detail() string }); ok {
		errorResp.Detail = err.Detail()
	}

	errorResp.Status = cmp.Or(errorResp.Status, http.StatusInternalServerError)
	w.WriteHeader(errorResp.Status)

	errorResp.Detail = cmp.Or(
		errorResp.Detail,
		http.StatusText(errorResp.Status),
		"An error occurred while processing the request",
	)

	slog.LogAttrs(r.Context(),
		slog.LevelError,
		"Error handling request",
		slog.String("path", request.GetPath(r)),
		slog.String("error", err.Error()))

	accept := strings.Split(r.Header.Get("Accept"), ",")

	for i := range accept {
		if j := strings.Index(accept[i], ";"); j != -1 {
			accept[i] = accept[i][:j]
		}
	}

	h := w.Header()
	h.Del("Content-Length")

	switch {
	case slices.Contains(accept, "application/json") || slices.Contains(accept, "*/*"):
		if err := json.NewEncoder(w).Encode(errorResp); err != nil {
			logWriteError(err)
			return
		}
		h.Set("Content-Type", "text/json; charset=utf-8")
		h.Set("X-Content-Type-Options", "nosniff")

	default: // default to plain text
		if _, err := w.Write([]byte(errorResp.Detail)); err != nil {
			logWriteError(err)
			return
		}
		h.Set("Content-Type", "text/plain; charset=utf-8")
		h.Set("X-Content-Type-Options", "nosniff")
	}
}
