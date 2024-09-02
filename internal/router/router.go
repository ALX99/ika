package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/proxy"
	"github.com/alx99/ika/middleware"
)

func MakeRouter(ctx context.Context, namespaces config.Namespaces) (http.Handler, error) {
	slog.Info("Building router", "middlewareCount", middleware.Len(), "namespaceCount", len(namespaces))

	mux := http.NewServeMux()

	for _, ns := range namespaces {
		log := slog.With(slog.String("namespace", ns.Name))
		transport := makeTransport(ns.Transport)
		p := proxy.NewProxy(transport)

		for pattern, path := range ns.Paths {
			middlewares := ns.Middlewares
			handler, err := getMiddleware(ctx, log, path, ns)
			if err != nil {
				return nil, err
			}

			routeHandler, err := p.Route(pattern, path.RewritePath, firstNonEmptyArr(path.Backends, ns.Backends))
			if err != nil {
				return nil, err
			}

			for _, pattern := range makeRoutePatterns(pattern, ns, path) {
				log.Info("Setting up path",
					"pattern", pattern,
					"namespace", ns.Name,
					"middlewares", slices.Collect(middlewares.Names()))

				mux.Handle(pattern, middleware.BindNamespace(ns.Name, handler(routeHandler)))
			}
		}
	}

	return mux, nil
}

// getMiddleware initializes the given middlewares and returns a handler that chains them for the given path and namespace
func getMiddleware(ctx context.Context, log *slog.Logger, path config.Path, ns config.Namespace) (func(http.Handler) http.Handler, error) {
	var handlers []func(http.Handler) http.Handler

	for mw := range path.MergedMiddlewares(ns) {
		log.Debug("Setting up middleware", "name", mw.Name)
		m, ok := middleware.Get(mw.Name)
		if !ok {
			return nil, fmt.Errorf("middleware %q has not been registered", mw.Name)
		}

		err := m.Setup(ctx, mw.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to setup middleware %q: %w", mw.Name, err)
		}

		handlers = append(handlers, m.Handle)
	}

	return chain(handlers), nil
}

// chain chains the given handlers in reverse order
func chain(handlers []func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for i := len(handlers) - 1; i >= 0; i-- {
			h = handlers[i](h)
		}
		return h
	}
}

func firstNonEmptyArr[T any](vs ...[]T) []T {
	for _, v := range vs {
		if len(v) > 0 {
			return v
		}
	}
	var empty []T
	return empty
}

func firstNonEmptyMap[T map[string]any](vs ...map[string]T) map[string]T {
	for _, v := range vs {
		if len(v) > 0 {
			return v
		}
	}
	var empty map[string]T
	return empty
}

func makeRoutePatterns(routePattern string, ns config.Namespace, route config.Path) []string {
	var patterns []string
	sb := strings.Builder{}

	if len(ns.Hosts) == 0 {
		if ns.DisableNamespacedPaths {
			return patterns // nothing to do
		}
	}

	if len(route.Methods) == 0 {

		if !ns.DisableNamespacedPaths {
			// Add namespaced route
			sb.WriteString("/")
			sb.WriteString(ns.Name)
			sb.WriteString(routePattern)
			patterns = append(patterns, sb.String())
		}

		for _, host := range ns.Hosts {
			sb.Reset()
			sb.WriteString(string(host))
			sb.WriteString(routePattern)
			// fmt.Printf("sb.String(): %v\n", sb.String())
			patterns = append(patterns, sb.String())
		}
	}

	for _, method := range route.Methods {
		sb.Reset()
		sb.WriteString(string(method))
		sb.WriteString(" ")

		if !ns.DisableNamespacedPaths {
			backup := sb.String()
			// Add namespaced route
			sb.WriteString("/")
			sb.WriteString(ns.Name)
			sb.WriteString(routePattern)
			patterns = append(patterns, sb.String())
			sb.Reset()
			sb.WriteString(backup)
		}

		backup := sb.String()
		for _, host := range ns.Hosts {
			sb.Reset()
			sb.WriteString(backup)

			sb.WriteString(string(host))
			sb.WriteString(routePattern)
			patterns = append(patterns, sb.String())
		}
	}

	return patterns
}

func makeTransport(cfg config.Transport) *http.Transport {
	return &http.Transport{
		DisableKeepAlives:      cfg.DisableKeepAlives,
		DisableCompression:     cfg.DisableCompression,
		MaxIdleConns:           cfg.MaxIdleConns,
		MaxIdleConnsPerHost:    cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:        cfg.MaxConnsPerHost,
		IdleConnTimeout:        cfg.IdleConnTimeout,
		ResponseHeaderTimeout:  cfg.ResponseHeaderTimeout,
		ExpectContinueTimeout:  cfg.ExpectContinueTimeout,
		MaxResponseHeaderBytes: cfg.MaxResponseHeaderBytes,
		WriteBufferSize:        cfg.WriteBufferSize,
		ReadBufferSize:         cfg.ReadBufferSize,
	}
}
