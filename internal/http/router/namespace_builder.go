package router

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/proxy"
	"github.com/alx99/ika/internal/http/request"
	"github.com/alx99/ika/internal/http/router/caramel"
	"github.com/alx99/ika/internal/http/router/chain"
	"github.com/alx99/ika/internal/teardown"
)

// nsBuilder handles the construction of a single namespace
type nsBuilder struct {
	name       string
	namespace  config.Namespace
	log        *slog.Logger
	factories  map[string]ika.PluginFactory
	teardowner teardown.Teardowner
}

func newNSBuilder(name string, ns config.Namespace, log *slog.Logger, factories map[string]ika.PluginFactory, mux *http.ServeMux) *nsBuilder {
	return &nsBuilder{
		name:       name,
		namespace:  ns,
		log:        log.With(slog.String("namespace", name)),
		factories:  factories,
		teardowner: make(teardown.Teardowner, 0),
	}
}

func (b *nsBuilder) build(ctx context.Context, mux *http.ServeMux) error {
	var transport http.RoundTripper = makeTransport(b.namespace.Transport)

	ictx := ika.InjectionContext{
		Namespace: b.name,
		Scope:     ika.ScopeNamespace,
		Logger:    b.log,
	}

	transport, err := b.setupTransport(ctx, ictx, transport)
	if err != nil {
		return errors.Join(err, b.teardowner.Teardown(ctx))
	}

	p, err := proxy.NewProxy(b.log, proxy.Config{
		Transport:  transport,
		Namespace:  b.name,
		BufferPool: newBufferPool(),
	})
	if err != nil {
		return errors.Join(err, b.teardowner.Teardown(ctx))
	}

	if err := b.buildRoutes(ctx, mux, p); err != nil {
		return errors.Join(err, b.teardowner.Teardown(ctx))
	}

	return nil
}

func (b *nsBuilder) buildRoutes(ctx context.Context, mux *http.ServeMux, p *proxy.Proxy) error {
	for _, mount := range b.namespace.Mounts {
		for pattern, route := range b.namespace.Routes {
			if err := b.buildRoute(ctx, mux, mount, pattern, route, p); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *nsBuilder) buildRoute(ctx context.Context, mux *http.ServeMux, mount, pattern string, route config.Route, p *proxy.Proxy) error {
	nsCtx := ika.InjectionContext{
		Namespace: b.name,
		Scope:     ika.ScopeNamespace,
		Logger:    b.log,
	}

	nsChain, err := b.makeChain(ctx, nsCtx,
		slices.Collect(b.namespace.Middlewares.Enabled()),
		slices.Collect(b.namespace.ReqModifiers.Enabled()),
		slices.Collect(b.namespace.Hooks.Enabled()),
	)
	if err != nil {
		return err
	}

	routeCtx := nsCtx
	routeCtx.Route = pattern
	routeCtx.Scope = ika.ScopeRoute

	routeChain, err := b.makeChain(ctx, routeCtx,
		slices.Collect(route.Middlewares.Enabled()),
		slices.Collect(route.ReqModifiers.Enabled()),
		nil,
	)
	if err != nil {
		return err
	}

	c := caramel.Wrap(mux).Mount(mount)
	patterns := b.generatePatterns(pattern, route.Methods)

	for _, pattern := range patterns {
		if b.shouldSkipPattern(pattern, mount) {
			continue
		}

		handlerChain := nsChain.Extend(routeChain).Then(p.WithPathTrim(mount))
		c.Handle(pattern, ika.ToHTTPHandler(handlerChain, buildErrHandler(b.log)))
	}

	return nil
}

func (b *nsBuilder) generatePatterns(pattern string, methods []config.Method) []string {
	if len(methods) == 0 {
		return []string{pattern}
	}

	patterns := make([]string, len(methods))
	for i, method := range methods {
		patterns[i] = string(method) + " " + pattern
	}
	return patterns
}

func (b *nsBuilder) shouldSkipPattern(pattern, mount string) bool {
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
	return pattern == "" && !strings.Contains(mount, "/")
}

func (b *nsBuilder) createPlugin(ctx context.Context, ictx ika.InjectionContext, cfg config.Plugin) (ika.Plugin, error) {
	ictx.Logger = ictx.Logger.With("plugin", cfg.Name)

	factory, ok := b.factories[cfg.Name]
	if !ok {
		return nil, fmt.Errorf("plugin %q not found", cfg.Name)
	}

	plugin, err := factory.New(ctx, ictx, cfg.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin %q: %w", cfg.Name, err)
	}

	b.teardowner.Add(plugin.Teardown)
	return plugin, nil
}

func (b *nsBuilder) setupTransport(ctx context.Context, ictx ika.InjectionContext, transport http.RoundTripper) (http.RoundTripper, error) {
	for _, cfg := range slices.Collect(b.namespace.Hooks.Enabled()) {
		plugin, err := b.createPlugin(ctx, ictx, cfg)
		if err != nil {
			return nil, err
		}

		hooker, ok := plugin.(ika.TripperHook)
		if !ok {
			// TODO issue here, the created plugin is never teared down
			continue // hooks does not have to implement every interface
		}

		transport, err = hooker.HookTripper(transport)
		if err != nil {
			return nil, err
		}
	}
	return transport, nil
}

func (b *nsBuilder) makeChain(ctx context.Context, ictx ika.InjectionContext, middlewares, reqModifiers, hooks config.Plugins) (chain.Chain, error) {
	ch := chain.New()

	// Add OnRequestHooks
	for _, cfg := range hooks {
		plugin, err := b.createPlugin(ctx, ictx, cfg)
		if err != nil {
			return chain.Chain{}, err
		}

		hooker, ok := plugin.(ika.OnRequestHook)
		if !ok {
			// TODO issue here, the created plugin is never teared down
			continue // hooks does not have to implement every interface
		}

		ch = ch.Append(chain.Constructor{
			Name:           cfg.Name,
			MiddlewareFunc: hooker.Handler,
		})
	}

	// Add RequestModifiers
	for _, cfg := range reqModifiers {
		plugin, err := b.createPlugin(ctx, ictx, cfg)
		if err != nil {
			return chain.Chain{}, err
		}

		modifier, ok := plugin.(ika.RequestModifier)
		if !ok {
			return chain.Chain{}, fmt.Errorf("plugin %q is not a RequestModifier", cfg.Name)
		}

		ch = ch.Append(chain.Constructor{
			Name: cfg.Name,
			MiddlewareFunc: func(next ika.Handler) ika.Handler {
				return ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					if err := modifier.ModifyRequest(r); err != nil {
						return err
					}
					return next.ServeHTTP(w, r)
				})
			},
		})
	}

	// Add Middlewares
	for _, cfg := range middlewares {
		plugin, err := b.createPlugin(ctx, ictx, cfg)
		if err != nil {
			return chain.Chain{}, err
		}

		mw, ok := plugin.(ika.Middleware)
		if !ok {
			return chain.Chain{}, fmt.Errorf("plugin %q is not a middleware", cfg.Name)
		}

		ch = ch.Append(chain.Constructor{
			Name:           cfg.Name,
			MiddlewareFunc: mw.Handler,
		})
	}

	return ch, nil
}

func (b *nsBuilder) teardown(ctx context.Context) error {
	return b.teardowner.Teardown(ctx)
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
