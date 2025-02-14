package accesslog

// https://pkg.go.dev/github.com/alx99/ika/plugins/accesslog

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/http/request"
	"github.com/alx99/ika/pluginutil"
	"github.com/felixge/httpsnoop"
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
	if err := pluginutil.UnmarshalCfg(config, &cfg); err != nil {
		return err
	}

	p.ikaPattern = ictx.Route
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
	var err error

	metrics := httpsnoop.CaptureMetricsFn(w,
		func(w http.ResponseWriter) { err = p.next.ServeHTTP(w, r) })

	attrs := []slog.Attr{
		slog.Group("request", p.makeReqAttrs(r)...),
		slog.Group("response",
			slog.Int64("duration", metrics.Duration.Milliseconds()),
			slog.Int("status", metrics.Code),
			slog.Int64("bytesWritten", metrics.Written),
		),
		slog.Group("ika",
			slog.String("pattern", r.Pattern),
		),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	p.log.LogAttrs(r.Context(), slog.LevelInfo, "access", attrs...)
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

	if p.cfg.RemoteAddr {
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
