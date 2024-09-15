package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/hook"
	"github.com/alx99/ika/internal/middleware"
	"github.com/alx99/ika/internal/proxy"
	pubMW "github.com/alx99/ika/middleware"
)

type Router struct {
	// mux is the underlying http.ServeMux
	mux      *http.ServeMux
	teardown []func(context.Context) error
}

type routePattern struct {
	// pattern is the route pattern
	pattern string
	// isNamespaced is true if the route pattern is namespaced
	isNamespaced bool
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Shutdown shuts down the router
func (r *Router) Shutdown(ctx context.Context) error {
	var err error
	for _, t := range r.teardown {
		if e := t(ctx); e != nil {
			err = errors.Join(err, e)
		}
	}
	return err
}

func MakeRouter(ctx context.Context, namespaces config.Namespaces, hookFacs hook.HookFactories) (*Router, error) {
	slog.Info("Building router", "namespaceCount", len(namespaces))
	r := &Router{
		mux: http.NewServeMux(),
	}

	for _, ns := range namespaces {
		log := slog.With(slog.String("namespace", ns.Name))
		var transport http.RoundTripper
		transport = makeTransport(ns.Transport)

		transport, teardown, err := hookFacs.ApplyTspHooks(ctx, ns.Hooks, transport)
		r.teardown = append(r.teardown, teardown)
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx))
		}

		for pattern, routeCfg := range ns.Paths {
			for _, route := range makeRoutes(pattern, ns, routeCfg) {
				p := proxy.NewProxy(proxy.Config{
					Transport:      transport,
					RoutePattern:   pattern,
					IsNamespaced:   route.isNamespaced,
					Namespace:      ns.Name,
					RewritePattern: routeCfg.RewritePath,
					Backends:       firstNonEmptyArr(routeCfg.Backends, ns.Backends),
				})

				handler, err := r.applyMiddlewares(ctx, log, hookFacs, p, routeCfg, ns)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}

				log.Debug("Setting up path",
					"pattern", route,
					"namespace", ns.Name,
					"middlewares", slices.Collect(ns.Middlewares.Names()))

				handler, teardown, err = hookFacs.ApplyFirstHandlerHook(ctx, ns.Hooks, handler)
				r.teardown = append(r.teardown, teardown)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}

				r.mux.Handle(route.pattern, pubMW.BindNamespace(ns.Name, handler))
			}
		}
	}

	return r, nil
}

// applyMiddlewares initializes the given middlewares and returns a handler that chains them for the given path and namespace
func (r *Router) applyMiddlewares(ctx context.Context, log *slog.Logger, hooks hook.HookFactories, handler http.Handler, path config.Path, ns config.Namespace) (http.Handler, error) {
	for mwConfig := range path.MergedMiddlewares(ns) {
		log.Debug("Setting up middleware", "name", mwConfig.Name)
		mw, err := middleware.Get(ctx, mwConfig.Name, handler)
		if err != nil {
			return nil, err
		}

		// Be nice to the user
		if mwConfig.Config == nil {
			mwConfig.Config = make(map[string]any)
		}
		err = mw.Setup(ctx, mwConfig.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup middleware %q: %w", mwConfig.Name, err)
		}
		handler = mw

		var teardown func(context.Context) error
		handler, teardown, err = hooks.ApplyMiddlewareHooks(ctx, ns.Hooks, mwConfig.Name, handler)
		r.teardown = append(r.teardown, teardown)
		if err != nil {
			return nil, err
		}
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

func makeRoutes(rp string, ns config.Namespace, route config.Path) []routePattern {
	var patterns []routePattern
	sb := strings.Builder{}
	isNamespaced := !ns.IsRoot()

	// if the routepattern is empty, it is impossible to route by hosts
	// since 'example.com' is not a valid route for example
	if rp == "" {
		ns.Hosts = []string{}
	}

	// writeNamespacedRoute writes the namespaced route if isRoot is false, otherwise it writes the route pattern
	writeNamespacedRoute := func(isRoot bool) {
		if isRoot {
			// Add namespaced route
			sb.WriteString("/")
			sb.WriteString(ns.Name)
			sb.WriteString(rp)
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
