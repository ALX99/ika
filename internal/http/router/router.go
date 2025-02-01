package router

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/proxy"
	"github.com/alx99/ika/internal/http/request"
	"github.com/alx99/ika/internal/http/router/caramel"
	"github.com/alx99/ika/internal/http/router/chain"
	"github.com/alx99/ika/internal/iplugin"
	"github.com/alx99/ika/internal/pool"
	"github.com/alx99/ika/internal/teardown"
)

type Router struct {
	tder teardown.Teardowner
	mux  *http.ServeMux
	cfg  config.Config
	opts config.Options
	log  *slog.Logger
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Shutdown shuts down the router
func (r *Router) Shutdown(ctx context.Context) error {
	return r.tder.Teardown(ctx)
}

func New(cfg config.Config, opts config.Options, log *slog.Logger) (*Router, error) {
	return &Router{
		mux:  http.NewServeMux(),
		cfg:  cfg,
		opts: opts,
		log:  log,
	}, nil
}

func (r *Router) Build(ctx context.Context) error {
	r.log.Info("Building router", "namespaceCount", len(r.cfg.Namespaces))

	for nsName, ns := range r.cfg.Namespaces {
		if err := r.buildNamespace(ctx, nsName, ns); err != nil {
			return err
		}
	}

	return nil
}

func (r *Router) buildNamespace(ctx context.Context, nsName string, ns config.Namespace) error {
	log := r.log.With(slog.String("namespace", nsName))
	var transport http.RoundTripper = makeTransport(ns.Transport)

	cache := iplugin.NewPluginCache(r.opts.Plugins)
	ictx := ika.InjectionContext{
		Namespace: nsName,
		Level:     ika.LevelNamespace,
		Logger:    log,
	}

	tripperHooks, teardown, err := iplugin.UsePlugins(
		ctx, ictx, cache, collectIters(ns.Hooks.Enabled()),
		iplugin.MakeTripperHooks, false,
	)
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}
	r.tder.Add(teardown)

	transport, err = tripperHooks(transport)
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}

	p, err := proxy.NewProxy(log, proxy.Config{
		Transport:  transport,
		Namespace:  nsName,
		BufferPool: pool.NewBufferPool(),
	})
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}

	nsChain, teardown, err := r.makePluginChain(ctx, ictx, cache,
		collectIters(ns.Middlewares.Enabled()),
		collectIters(ns.ReqModifiers.Enabled()),
		collectIters(ns.Hooks.Enabled()),
	)
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}
	r.tder.Add(teardown)

	for _, mount := range ns.Mounts {
		c := caramel.Wrap(r.mux).Mount(mount)

		for pattern, path := range ns.Paths {
			ictx := ika.InjectionContext{
				Namespace:   nsName,
				PathPattern: pattern,
				Level:       ika.LevelPath,
				Logger:      log,
			}

			cache = iplugin.NewPluginCache(r.opts.Plugins)
			pathChain, teardown, err := r.makePluginChain(ctx, ictx, cache,
				collectIters(path.Middlewares.Enabled()),
				collectIters(path.ReqModifiers.Enabled()),
				nil,
			)
			if err != nil {
				return errors.Join(err, r.tder.Teardown(ctx))
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
				if pattern == "" && !strings.Contains(mount, "/") {
					// This is a special scenario where the namespace
					// path does not contain a '/'. In other words,
					// it must mean that it is a host.

					// since [HOST] alone is not a valid pattern
					// we ignore this specific scenario to allow
					// users to register path like:
					// mounts: ["/example"]
					// path: ""
					// to prevent redirection when /example is accessed
					// alone
					continue
				}

				nsChain := nsChain.Extend(pathChain).Then(p.WithPathTrim(mount))
				c.Handle(pattern, ika.ToHTTPHandler(nsChain, buildErrHandler(log)))
			}
		}
	}

	return nil
}

func (r *Router) makePluginChain(ctx context.Context, ictx ika.InjectionContext, setupper *iplugin.PluginCache, middlewares, reqModifiers, hooks config.Plugins) (chain.Chain, teardown.TeardownFunc, error) {
	var tder teardown.Teardowner

	mwChain, teardown, err := iplugin.UsePlugins(
		ctx, ictx, setupper, middlewares,
		iplugin.ChainFromMiddlewares, true)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	reqModChain, teardown, err := iplugin.UsePlugins(
		ctx, ictx, setupper, reqModifiers,
		iplugin.ChainFromReqModifiers, true)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	onReqChain, teardown, err := iplugin.UsePlugins(
		ctx, ictx, setupper, hooks,
		iplugin.ChainOnRequestHooks, false)
	if err != nil {
		return chain.Chain{}, tder.Teardown, errors.Join(err, tder.Teardown(ctx))
	}
	tder.Add(teardown)

	ch := chain.New().Extend(onReqChain).Extend(reqModChain).Extend(mwChain)

	return ch, teardown, nil
}

func makeTransport(cfg config.Transport) *http.Transport {
	d := net.Dialer{
		Timeout:       cfg.Dialer.Timeout.Dur(),
		FallbackDelay: cfg.Dialer.FallbackDelay.Dur(),
		KeepAliveConfig: net.KeepAliveConfig{
			Enable:   cfg.Dialer.KeepAliveConfig.Enable,
			Idle:     cfg.Dialer.KeepAliveConfig.Idle.Dur(),
			Interval: cfg.Dialer.KeepAliveConfig.Interval.Dur(),
			Count:    cfg.Dialer.KeepAliveConfig.Count,
		},
	}
	return &http.Transport{
		DialContext:            d.DialContext,
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

func buildErrHandler(log *slog.Logger) ika.ErrorHandler {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		log.LogAttrs(r.Context(),
			slog.LevelError,
			"Error handling request",
			slog.String("path", request.GetPath(r)),
			slog.String("error", err.Error()))
		http.Error(w, "failed to handle request", http.StatusInternalServerError)
	}
}
