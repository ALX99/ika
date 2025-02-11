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
	tder  teardown.Teardowner
	mux   *http.ServeMux
	cfg   config.Config
	opts  config.Options
	log   *slog.Logger
	cache *iplugin.PluginCache
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Shutdown shuts down the router
func (r *Router) Shutdown(ctx context.Context) error {
	r.tder.Add(r.cache.Teardown)
	return r.tder.Teardown(ctx)
}

func New(cfg config.Config, opts config.Options, log *slog.Logger) (*Router, error) {
	return &Router{
		mux:   http.NewServeMux(),
		cfg:   cfg,
		opts:  opts,
		log:   log,
		cache: iplugin.NewPluginCache(opts.Plugins),
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

	ictx := ika.InjectionContext{
		Namespace: nsName,
		Level:     ika.LevelNamespace,
		Logger:    log,
	}

	plugins, err := r.cache.GetPlugins(ctx, ictx, collectIters(ns.Hooks.Enabled()))
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}

	transport, err = iplugin.MakeTripperHooks(plugins)(transport)
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

	nsChain, err := r.makePluginChain(ctx, ictx,
		collectIters(ns.Middlewares.Enabled()),
		collectIters(ns.ReqModifiers.Enabled()),
		collectIters(ns.Hooks.Enabled()),
	)
	if err != nil {
		return errors.Join(err, r.tder.Teardown(ctx))
	}

	for _, mount := range ns.Mounts {
		c := caramel.Wrap(r.mux).Mount(mount)

		for pattern, path := range ns.Paths {
			ictx := ika.InjectionContext{
				Namespace:   nsName,
				PathPattern: pattern,
				Level:       ika.LevelPath,
				Logger:      log,
			}

			pathChain, err := r.makePluginChain(ctx, ictx,
				collectIters(path.Middlewares.Enabled()),
				collectIters(path.ReqModifiers.Enabled()),
				nil,
			)
			if err != nil {
				return errors.Join(err, r.tder.Teardown(ctx))
			}

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

func (r *Router) makePluginChain(ctx context.Context, ictx ika.InjectionContext, middlewares, reqModifiers, hooks config.Plugins) (chain.Chain, error) {
	mwPlugins, err := r.cache.GetPlugins(ctx, ictx, middlewares)
	if err != nil {
		return chain.Chain{}, err
	}

	mwChain, err := iplugin.ChainFromMiddlewares(mwPlugins)
	if err != nil {
		return chain.Chain{}, err
	}

	reqModPlugins, err := r.cache.GetPlugins(ctx, ictx, reqModifiers)
	if err != nil {
		return chain.Chain{}, err
	}

	reqModChain, err := iplugin.ChainFromReqModifiers(reqModPlugins)
	if err != nil {
		return chain.Chain{}, err
	}

	onReqPlugins, err := r.cache.GetPlugins(ctx, ictx, hooks)
	if err != nil {
		return chain.Chain{}, err
	}

	onReqChain, err := iplugin.MakeOnRequestHooks(onReqPlugins)
	if err != nil {
		return chain.Chain{}, err
	}

	ch := chain.New().Extend(onReqChain).Extend(reqModChain).Extend(mwChain)

	return ch, nil
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
		ika.DefaultErrorHandler(w, r, err)
	}
}
