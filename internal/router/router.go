package router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/iplugin"
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
func (r *Router) Shutdown(ctx context.Context, err error) error {
	for _, t := range r.teardown {
		err = errors.Join(err, t(ctx))
	}
	return err
}

func MakeRouter(ctx context.Context, cfg config.Config, opts config.Options) (*Router, error) {
	slog.Info("Building router", "namespaceCount", len(cfg.Namespaces))
	r := &Router{
		mux: http.NewServeMux(),
	}

	for nsName, ns := range cfg.Namespaces {
		log := slog.With(slog.String("namespace", nsName))
		var transport http.RoundTripper = makeTransport(ns.Transport)

		setupper := iplugin.NewSetupper(opts.Plugins)
		iCtx := plugin.InjectionContext{
			Namespace: nsName,
			Level:     plugin.LevelNamespace,
		}

		wrapTransport, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, collectIters(ns.Hooks.Enabled()), iplugin.MakeTransportWrapper)
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx, err))
		}
		r.teardown = append(r.teardown, teardown)

		transport = wrapTransport(transport)

		p, err := proxy.NewProxy(proxy.Config{
			Transport:  transport,
			Namespace:  nsName,
			BufferPool: &pool.BufferPool{Pool: bytebufferpool.Pool{}},
		})
		if err != nil {
			return nil, errors.Join(err, r.Shutdown(ctx, err))
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

				middlewares := collectIters(ns.Middlewares.Enabled(), path.Middlewares.Enabled())
				mwChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, middlewares, iplugin.ChainFromMiddlewares)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx, err))
				}
				r.teardown = append(r.teardown, teardown)

				reqModifiers := collectIters(ns.ReqModifiers.Enabled(), path.ReqModifiers.Enabled())
				reqModChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, reqModifiers, iplugin.ChainFromReqModifiers)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx, err))
				}
				r.teardown = append(r.teardown, teardown)

				firstHandlerChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, collectIters(ns.Hooks.Enabled()),
					iplugin.ChainFirstHandler)
				if err != nil {
					return nil, errors.Join(err, r.Shutdown(ctx, err))
				}
				r.teardown = append(r.teardown, teardown)

				ch := chain.New().Extend(firstHandlerChain).Extend(reqModChain).Extend(mwChain)

				log.Debug("Path registered",
					"pattern", route.pattern,
					"middlewares", slices.Collect(config.Plugins(middlewares).Names()),
					"reqModifiers", slices.Collect(config.Plugins(reqModifiers).Names()),
				)

				r.mux.Handle(route.pattern, pubMW.BindMetadata(pubMW.Metadata{
					Namespace:      nsName,
					Route:          pattern,
					GeneratedRoute: route.pattern,
				}, ch.Then(plugin.WrapHTTPHandler(p))))
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
		DisableKeepAlives:      cfg.DisableKeepAlives,
		DisableCompression:     cfg.DisableCompression,
		MaxIdleConns:           cfg.MaxIdleConns,
		MaxIdleConnsPerHost:    cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:        cfg.MaxConnsPerHost,
		IdleConnTimeout:        cfg.IdleConnTimeout.Dur(),
		ResponseHeaderTimeout:  cfg.ResponseHeaderTimeout.Dur(),
		ExpectContinueTimeout:  cfg.ExpectContinueTimeout.Dur(),
		MaxResponseHeaderBytes: cfg.MaxResponseHeaderBytes,
		WriteBufferSize:        cfg.WriteBufferSize,
		ReadBufferSize:         cfg.ReadBufferSize,
	}
}
