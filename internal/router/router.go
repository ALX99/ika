package router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/pool"
	"github.com/alx99/ika/internal/proxy"
	"github.com/alx99/ika/internal/router/chain"
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
		log := slog.With(slog.String("namespace", nsName))
		var transport http.RoundTripper = makeTransport(ns.Transport)

		transport, teardown, err := cfg.WrapTransport(ctx, ns.Plugins, transport)
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx))
		}
		r.teardown = append(r.teardown, teardown)

		p, err := proxy.NewProxy(proxy.Config{
			Transport:  transport,
			Namespace:  nsName,
			BufferPool: &pool.BufferPool{Pool: bytebufferpool.Pool{}},
		})
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx))
		}

		// TODO don't explode namespace level request modifiers / middlewares into paths
		// create multiple subrouters for each namespace in the future
		// Then plugins will be able to run on both namespace and path level
		for pattern, path := range ns.Paths {
			for _, route := range makeRoutes(pattern, nsName, ns, path) {
				iCtx := plugin.InjectionContext{
					Namespace:   nsName,
					PathPattern: pattern,
					Level:       plugin.LevelPath,
				}

				mwChain, teardown, err := makePluginChain(ctx, iCtx, collectIters(ns.Middlewares.Enabled(), path.Middlewares.Enabled()), cfg.PluginFacs2, handlerFromMiddlewares)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}
				r.teardown = append(r.teardown, teardown)

				reqModChain, teardown, err := makePluginChain(ctx, iCtx, collectIters(ns.ReqModifiers.Enabled(), path.ReqModifiers.Enabled()),
					cfg.PluginFacs2, handlerFromRequestModifiers)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}
				r.teardown = append(r.teardown, teardown)

				ch := chain.New().Extend(reqModChain).Extend(mwChain)

				handler, teardown, err := cfg.WrapFirstHandler(ctx, ns.Plugins, ch.Then(plugin.WrapHTTPHandler(p)))
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx))
				}
				r.teardown = append(r.teardown, teardown)

				log.Debug("Path registered",
					"pattern", route.pattern,
					"middlewares", slices.Collect(path.Middlewares.Names()),
					"reqModifiers", slices.Collect(path.ReqModifiers.Names()),
				)

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
