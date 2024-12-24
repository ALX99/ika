package router

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/iplugin"
	"github.com/alx99/ika/internal/pool"
	"github.com/alx99/ika/internal/proxy"
	"github.com/alx99/ika/internal/router/chain"
	"github.com/alx99/ika/internal/teardown"
	"github.com/alx99/ika/plugin"
	"github.com/valyala/bytebufferpool"
)

type Router struct {
	// mux is the underlying http.ServeMux
	nsRouter namespacedRouter
	tder     teardown.Teardowner
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.nsRouter.ServeHTTP(w, req)
}

// Shutdown shuts down the router
func (r *Router) Shutdown(ctx context.Context) error {
	return r.tder.Teardown(ctx)
}

func MakeRouter(ctx context.Context, cfg config.Config, opts config.Options) (*Router, error) {
	log := slog.With(slog.String("module", "router"))
	log.Info("Building router", "namespaceCount", len(cfg.Namespaces))

	r := &Router{}

	nsRouter := namespacedRouter{
		namespaces: make(map[string]namespace),
		router:     router{mux: http.NewServeMux()},
		log:        log,
	}

	for nsName, ns := range cfg.Namespaces {
		log = log.With(slog.String("namespace", nsName))
		var transport http.RoundTripper = makeTransport(ns.Transport)

		setupper := iplugin.NewSetupper(opts.Plugins)
		iCtx := plugin.InjectionContext{
			Namespace: nsName,
			Level:     plugin.LevelNamespace,
			Logger:    slog.Default().With(slog.String("namespace", nsName)),
		}

		wrapTransport, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, collectIters(ns.Hooks.Enabled()), iplugin.MakeTransportWrapper)
		if err != nil {
			return nil, errors.Join(err, r.tder.Teardown(ctx))
		}
		r.tder.Add(teardown)

		transport = wrapTransport(transport)

		p, err := proxy.NewProxy(proxy.Config{
			Transport:  transport,
			Namespace:  nsName,
			BufferPool: &pool.BufferPool{Pool: bytebufferpool.Pool{}},
		})
		if err != nil {
			return nil, errors.Join(err, r.tder.Teardown(ctx))
		}

		nsRouter.addNamespace(namespace{
			name:       nsName,
			nsPaths:    ns.NSPaths,
			mux:        http.NewServeMux(),
			addedPaths: make(map[string]struct{}),
		})

		setupper = iplugin.NewSetupper(opts.Plugins)
		nsChain, teardown, err := r.makePluginChain(ctx, iCtx, setupper,
			collectIters(ns.Middlewares.Enabled()),
			collectIters(ns.ReqModifiers.Enabled()),
			collectIters(ns.Hooks.Enabled()),
		)
		if err != nil {
			return nil, errors.Join(err, r.tder.Teardown(ctx))
		}
		r.tder.Add(teardown)

		for pattern, path := range ns.Paths {
			iCtx := plugin.InjectionContext{
				Namespace:   nsName,
				PathPattern: pattern,
				Level:       plugin.LevelPath,
				Logger:      slog.Default().With(slog.String("namespace", nsName)),
			}

			setupper = iplugin.NewSetupper(opts.Plugins)
			pathChain, teardown, err := r.makePluginChain(ctx, iCtx, setupper,
				collectIters(path.Middlewares.Enabled()),
				collectIters(path.ReqModifiers.Enabled()),
				nil,
			)
			if err != nil {
				return nil, errors.Join(err, r.tder.Teardown(ctx))
			}
			r.tder.Add(teardown)

			var patterns []string
			if len(path.Methods) == 0 {
				patterns = append(patterns, pattern)
			} else {
				for _, method := range path.Methods {
					patterns = append(patterns, string(method)+" "+pattern)
				}
			}

			for _, pattern := range patterns {
				nsRouter.addNamespacePath(nsName, pattern,
					plugin.WrapErrHandler(nsChain.Extend(pathChain).Then(p), defaultErrHandler))
			}
		}

	}

	r.nsRouter = nsRouter

	return r, nil
}

func (r *Router) makePluginChain(ctx context.Context, iCtx plugin.InjectionContext, setupper *iplugin.PluginSetupper, middlewares, reqModifiers, hooks config.Plugins) (chain.Chain, teardown.TeardownFunc, error) {
	var tder teardown.Teardowner

	mwChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, middlewares, iplugin.ChainFromMiddlewares)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	reqModChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, reqModifiers, iplugin.ChainFromReqModifiers)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	firstHandlerChain, teardown, err := iplugin.UsePlugins(ctx, iCtx, setupper, hooks, iplugin.ChainFirstHandler)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	ch := chain.New().Extend(firstHandlerChain).Extend(reqModChain).Extend(mwChain)

	return ch, teardown, nil
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

func defaultErrHandler(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("Error handling request", "err", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
