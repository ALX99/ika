package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/middleware"
	"github.com/alx99/ika/internal/proxy"
	pubMW "github.com/alx99/ika/middleware"
)

type routePattern struct {
	// pattern is the route pattern
	pattern string
	// isNamespaced is true if the route pattern is namespaced
	isNamespaced bool
}

func MakeRouter(ctx context.Context, namespaces config.Namespaces) (http.Handler, error) {
	slog.Info("Building router", "namespaceCount", len(namespaces))

	mux := http.NewServeMux()

	for _, ns := range namespaces {
		log := slog.With(slog.String("namespace", ns.Name))
		transport := makeTransport(ns.Transport)
		p := proxy.NewProxy(transport)

		for pattern, routeCfg := range ns.Paths {
			for _, route := range makeRoutes(pattern, ns, routeCfg) {

				rewritePath := routeCfg.RewritePath.V
				// If the route is namespaced and the rewrite path is empty, set the rewrite path to the route pattern
				// This ensures that the namespaced prefix is stripped from the request path
				if route.isNamespaced && routeCfg.RewritePath.V == "" {
					rewritePath = routeCfg.RewritePath.Or(strings.TrimPrefix(route.pattern, "/"+ns.Name))
				}

				proxyHandler, err := p.GetHandler(pattern, route.isNamespaced, rewritePath, firstNonEmptyArr(routeCfg.Backends, ns.Backends))
				if err != nil {
					return nil, err
				}

				log.Debug("Setting up path",
					"pattern", route,
					"namespace", ns.Name,
					"middlewares", slices.Collect(ns.Middlewares.Names()))

				handler, err := applyMiddlewares(ctx, log, proxyHandler, routeCfg, ns)
				if err != nil {
					return nil, err
				}

				mux.Handle(route.pattern, pubMW.BindNamespace(ns.Name, handler))
			}
		}
	}

	return mux, nil
}

// applyMiddlewares initializes the given middlewares and returns a handler that chains them for the given path and namespace
func applyMiddlewares(ctx context.Context, log *slog.Logger, handler http.Handler, path config.Path, ns config.Namespace) (http.Handler, error) {
	for mwConfig := range path.MergedMiddlewares(ns) {
		log.Debug("Setting up middleware", "name", mwConfig.Name)
		mw, err := middleware.Get(ctx, mwConfig.Name, handler)
		if err != nil {
			return nil, err
		}

		err = mw.Setup(ctx, mwConfig.Args)
		if err != nil {
			return nil, fmt.Errorf("failed to setup middleware %q: %w", mwConfig.Name, err)
		}

		handler = mw
	}

	return handler, nil
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

func makeRoutes(rp string, ns config.Namespace, route config.Path) []routePattern {
	var patterns []routePattern
	sb := strings.Builder{}

	// writeNamespacedRoute writes the namespaced route if isRoot is false, otherwise it writes the route pattern
	writeNamespacedRoute := func(isRoot bool) {
		if isRoot {
			// Add namespaced route
			sb.WriteString("/")
			sb.WriteString(ns.Name)
			sb.WriteString(rp)
			isRoot = true
		} else {
			sb.WriteString(rp)
		}
	}

	if len(ns.Hosts) == 0 {
		if ns.DisableNamespacedPaths.V {
			return patterns // nothing to do
		}
	}

	if len(route.Methods) == 0 {
		if !ns.DisableNamespacedPaths.V {
			isNamespaced := !ns.IsRoot()
			writeNamespacedRoute(isNamespaced)
			patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: isNamespaced})
		}

		for _, host := range ns.Hosts {
			sb.Reset()
			sb.WriteString(string(host))
			sb.WriteString(rp)
			patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: false})
		}
	}

	for _, method := range route.Methods {
		sb.Reset()
		sb.WriteString(string(method))
		sb.WriteString(" ")

		if !ns.DisableNamespacedPaths.V {
			backup := sb.String()
			isNamespaced := !ns.IsRoot()
			writeNamespacedRoute(isNamespaced)
			patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: isNamespaced})
			sb.Reset()
			sb.WriteString(backup)
		}

		backup := sb.String()
		for _, host := range ns.Hosts {
			sb.Reset()
			sb.WriteString(backup)

			sb.WriteString(string(host))
			sb.WriteString(rp)
			patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: false})
		}
	}

	return patterns
}

func makeTransport(cfg config.Transport) *http.Transport {
	return &http.Transport{
		DisableKeepAlives:      cfg.DisableKeepAlives.V,
		DisableCompression:     cfg.DisableCompression.V,
		MaxIdleConns:           cfg.MaxIdleConns.V,
		MaxIdleConnsPerHost:    cfg.MaxIdleConnsPerHost.V,
		MaxConnsPerHost:        cfg.MaxConnsPerHost.V,
		IdleConnTimeout:        cfg.IdleConnTimeout.V,
		ResponseHeaderTimeout:  cfg.ResponseHeaderTimeout.V,
		ExpectContinueTimeout:  cfg.ExpectContinueTimeout.V,
		MaxResponseHeaderBytes: cfg.MaxResponseHeaderBytes.V,
		WriteBufferSize:        cfg.WriteBufferSize.V,
		ReadBufferSize:         cfg.ReadBufferSize.V,
	}
}
