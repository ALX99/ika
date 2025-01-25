package accesslog

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/http/request"
)

type Plugin struct {
	cfg pConfig

	ikaPattern     string
	includeHeaders bool
	next           ika.Handler
	log            *slog.Logger
}

func (Plugin) New(_ context.Context, _ ika.InjectionContext) (ika.Plugin, error) {
	return &Plugin{}, nil
}

func (Plugin) Name() string {
	return "access-log"
}

func (p *Plugin) Setup(ctx context.Context, ictx ika.InjectionContext, config map[string]any) error {
	cfg := pConfig{}
	if err := toStruct(config, &cfg); err != nil {
		return err
	}

	p.ikaPattern = ictx.PathPattern
	p.log = ictx.Logger
	p.cfg = cfg
	p.includeHeaders = len(cfg.Headers) > 0

	return nil
}

func (p *Plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	now := time.Now()
	err := p.next.ServeHTTP(w, r)
	end := time.Now()

	attrs := []slog.Attr{
		slog.Group("request", p.makeReqAttrs(r)...),
		slog.Group("response",
			slog.Int64("duration", end.Sub(now).Milliseconds()),
		),
		slog.Group("ika",
			slog.String("pattern", r.Pattern),
		),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	p.log.LogAttrs(r.Context(), slog.LevelInfo, "endpoint access", attrs...)
	return err
}

func (Plugin) Teardown(context.Context) error { return nil }

func (p *Plugin) makeReqAttrs(r *http.Request) []any {
	requestAttrs := []any{
		slog.String("method", r.Method),
		slog.String("path", request.GetPath(r)),
		slog.String("host", r.Host),
		slog.String("pattern", r.Pattern),
	}

	if p.cfg.IncludeRemoteAddr {
		requestAttrs = append(requestAttrs, slog.String("remoteAddr", r.RemoteAddr))
	}

	if p.includeHeaders {
		attrs := make([]any, 0, len(p.cfg.Headers))
		for _, key := range p.cfg.Headers {
			if val := r.Header.Get(key); val != "" {
				attrs = append(attrs, slog.String(key, val))
			}
		}
		requestAttrs = append(requestAttrs, slog.Group("headers", attrs...))
	}

	return requestAttrs
}

var (
	_ ika.Middleware    = &Plugin{}
	_ ika.PluginFactory = &Plugin{}
)
