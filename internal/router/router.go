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
	"github.com/alx99/ika/internal/middleware"
	"github.com/alx99/ika/internal/pool"
	"github.com/alx99/ika/internal/proxy"
	pubMW "github.com/alx99/ika/middleware"
	"github.com/valyala/bytebufferpool"
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

func MakeRouter(ctx context.Context, cfg config.Config) (*Router, error) {
	slog.Info("Building router", "namespaceCount", len(cfg.Namespaces))
	r := &Router{
		mux: http.NewServeMux(),
	}

	for nsName, ns := range cfg.Namespaces {
		bPool := &pool.BufferPool{Pool: bytebufferpool.Pool{}}
		log := slog.With(slog.String("namespace", nsName))
		var transport http.RoundTripper
		transport = makeTransport(ns.Transport)

		transport, teardown, err := cfg.WrapTransport(ctx, ns.Plugins, transport)
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx))
		}
		r.teardown = append(r.teardown, teardown)

		for pattern, routeCfg := range ns.Paths {
			for _, route := range makeRoutes(pattern, nsName, ns, routeCfg) {
				p := proxy.NewProxy(proxy.Config{
					Transport:      transport,
					RoutePattern:   pattern,
					IsNamespaced:   route.isNamespaced,
					Namespace:      nsName,
					RewritePattern: routeCfg.RewritePath,
					Backends:       firstNonEmptyArr(routeCfg.Backends, ns.Backends),
					BufferPool:     bPool,
				})

				handler, err := r.applyMiddlewares(ctx, log, cfg, p, routeCfg, ns)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}

				log.Debug("Setting up path",
					"pattern", route.pattern,
					"middlewares", slices.Collect(ns.Middlewares.Names()))

				handler, teardown, err = cfg.WrapFirstHandler(ctx, ns.Plugins, handler)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}
				r.teardown = append(r.teardown, teardown)

				r.mux.Handle(route.pattern, pubMW.BindMetadata(pubMW.Metadata{
					Namespace:      nsName,
					Route:          pattern,
					GeneratedRoute: route.pattern,
				}, handler))
			}
		}
	}

	return r, nil
}

// applyMiddlewares initializes the given middlewares and returns a handler that chains them for the given path and namespace
func (r *Router) applyMiddlewares(ctx context.Context, log *slog.Logger, cfg config.Config, handler http.Handler, path config.Path, ns config.Namespace) (http.Handler, error) {
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
		handler, teardown, err = cfg.WrapMiddleware(ctx, ns.Plugins, mwConfig.Name, handler)
		if err != nil {
			return nil, err
		}
		r.teardown = append(r.teardown, teardown)
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

func makeRoutes(rp string, nsName string, ns config.Namespace, route config.Path) []routePattern {
	var patterns []routePattern
	sb := strings.Builder{}
	isRoot := nsName == "root"
	isHost := nsName[0] != '/' && !isRoot
	isNamespaced := !isHost && !isRoot

	// impossible to register a route like this
	if isHost && rp == "" {
		return patterns
	}

	// writeRoute writes the namespaced route if isRoot is false, otherwise it writes the route pattern
	writeRoute := func() {
		if isNamespaced {
			sb.WriteString(nsName)
			sb.WriteString(rp)
		} else if isHost {
			sb.WriteString(nsName)
			sb.WriteString(rp)
		} else {
			// must be root namespace
			sb.WriteString(rp)
		}
	}

	if len(route.Methods) == 0 {
		writeRoute()
		patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: isNamespaced})
		return patterns
	}

	for _, method := range route.Methods {
		sb.WriteString(string(method))
		sb.WriteString(" ")

		writeRoute()
		patterns = append(patterns, routePattern{pattern: sb.String(), isNamespaced: isNamespaced})
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
