package router

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/proxy"
	"github.com/alx99/ika/middleware"
)

type routerMaker struct {
	mwsInitialized map[string]bool
}

func MakeRouter(ctx context.Context, namespaces config.Namespaces) (http.Handler, error) {
	slog.Info("Building router", "middlewareCount", middleware.Len(), "namespaceCount", len(namespaces))

	rm := routerMaker{mwsInitialized: make(map[string]bool)}
	mux := http.NewServeMux()

	for name, ns := range namespaces {
		log := slog.With(slog.String("namespace", name))
		p := proxy.NewProxy(ns.Transport)

		for pattern, path := range ns.Paths {
			log = log.With(slog.String("pattern", pattern))
			middlewares := firstNonEmptyMap(path.Middlewares, ns.Middlewares)
			handler, err := rm.getMiddleware(ctx, log, middlewares)
			if err != nil {
				return nil, err
			}

			routeHandler, err := p.Route(firstNonEmptyArr(path.Backends, ns.Backends))
			if err != nil {
				return nil, err
			}

			for _, pattern := range makeRoutePatterns(pattern, name, ns, path) {
				log.Info("Setting up path",
					"pattern", pattern,
					"namespace", name,
					"middlewares", slices.Collect(maps.Keys(middlewares)))

				mux.Handle(pattern, bindNamespace(name, handler(routeHandler)))
			}
		}
	}

	return mux, nil
}

// getMiddleware initializes the given middlewares and returns a handler that chains them
func (rm routerMaker) getMiddleware(ctx context.Context, log *slog.Logger, mws config.Middlewares) (func(http.Handler) http.Handler, error) {
	var handlers []func(http.Handler) http.Handler
	for name, cfg := range mws {
		if rm.mwsInitialized[name] {
			continue
		}

		log.Debug("Setting up middleware", "name", name)
		m, ok := middleware.Get(name)
		if !ok {
			return nil, fmt.Errorf("middleware %q has not been registered", name)
		}
		rm.mwsInitialized[name] = true

		err := m.Setup(ctx, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to setup middleware %q: %w", name, err)
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

func makeRoutePatterns(routePattern, nsName string, ns config.Namespace, route config.Path) []string {
	var patterns []string
	sb := strings.Builder{}
	for _, method := range route.Methods {
		sb.Reset()
		sb.WriteString(string(method))
		sb.WriteString(" ")

		if len(ns.Hosts) == 0 {
			if ns.DisableNamespacedPaths {
				continue // nothing to do
			}

			sb.WriteString("/")
			sb.WriteString(nsName)
			sb.WriteString(routePattern)
			patterns = append(patterns, sb.String())
			continue
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
