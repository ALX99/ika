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
	"github.com/alx99/ika/internal/pool"
	"github.com/alx99/ika/internal/proxy"
	pubMW "github.com/alx99/ika/middleware"
	"github.com/alx99/ika/plugin"
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

		for pattern, path := range ns.Paths {
			for _, route := range makeRoutes(pattern, nsName, ns, path) {
				p, err := proxy.NewProxy(proxy.Config{
					Transport:  transport,
					Namespace:  nsName,
					Backends:   firstNonEmptyArr(path.Redirect.Backends, ns.Backends),
					BufferPool: bPool,
				})
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}

				handler, err := makeMiddlewaresHandler(ctx, p, path, pattern, ns, nsName, cfg.RequestModifiers)
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

				handler, err = makeRequestModifierHandler(ctx, handler, path, pattern, ns, nsName, cfg.RequestModifiers)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}

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

func makeRequestModifierHandler(ctx context.Context,
	next http.Handler,
	path config.Path,
	pathPattern string,
	ns config.Namespace,
	nsName string,
	requestModifiersFaqs map[string]plugin.NFactory,
) (http.Handler, error) {
	requestModifiers := []plugin.RequestModifier{}
	for pluginCfg := range path.Plugins.Enabled() {
		p, err := requestModifiersFaqs[pluginCfg.Name].New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin %q: %w", pluginCfg.Name, err)
		}

		// Does not support modifying requests
		if !slices.Contains(p.Capabilities(), plugin.CapModifyRequests) {
			continue
		}

		if !slices.Contains(p.InjectionLevels(), plugin.PathLevel) {
			return nil, fmt.Errorf("plugin %q does not support path level injection", p.Name())
		}

		err = p.Setup(ctx, plugin.InjectionContext{
			Namespace:   nsName,
			PathPattern: pathPattern,
			Level:       plugin.PathLevel,
		}, pluginCfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup plugin %q: %w", p.Name(), err)
		}

		rm, ok := p.(plugin.RequestModifier)
		if !ok {
			return nil, fmt.Errorf("plugin %q does not implement RequestModifier", p.Name())
		}

		requestModifiers = append(requestModifiers, rm)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		for _, rm := range requestModifiers {
			r, err = rm.ModifyRequest(ctx, r)
			if err != nil {
				panic(err) // todo
			}
		}
		next.ServeHTTP(w, r)
	}), nil
}

func makeMiddlewaresHandler(ctx context.Context,
	next http.Handler,
	path config.Path,
	pathPattern string,
	ns config.Namespace,
	nsName string,
	requestModifiersFaqs map[string]plugin.NFactory,
) (http.Handler, error) {
	middlewares := []plugin.Middleware{}
	for pluginCfg := range path.Plugins.Enabled() {
		p, err := requestModifiersFaqs[pluginCfg.Name].New(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create plugin %q: %w", pluginCfg.Name, err)
		}

		// Does not support middleware
		if !slices.Contains(p.Capabilities(), plugin.CapMiddleware) {
			continue
		}

		if !slices.Contains(p.InjectionLevels(), plugin.PathLevel) {
			return nil, fmt.Errorf("plugin %q does not support path level injection", p.Name())
		}

		err = p.Setup(ctx, plugin.InjectionContext{
			Namespace:   nsName,
			PathPattern: pathPattern,
			Level:       plugin.PathLevel,
		}, pluginCfg.Config)
		if err != nil {
			return nil, fmt.Errorf("failed to setup plugin %q: %w", p.Name(), err)
		}

		mw, ok := p.(plugin.Middleware)
		if !ok {
			return nil, fmt.Errorf("plugin %q does not implement Middleware", p.Name())
		}

		middlewares = append(middlewares, mw)
	}

	if len(middlewares) == 0 {
		return next, nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var nextE plugin.ErrHandler = plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			next.ServeHTTP(w, r)
			return nil
		})
		for _, middleware := range middlewares {
			var err error
			nextE, err = middleware.Handler(ctx, nextE)
			if err != nil {
				panic(err) // todo
			}
		}
		nextE.ServeHTTP(w, r)
	}), nil
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
