package accesslog

// https://pkg.go.dev/github.com/alx99/ika/plugins/accesslog

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/http/request"
	"github.com/alx99/ika/pluginutil"
	"github.com/felixge/httpsnoop"
)

type plugin struct {
	cfg pConfig

	ikaPattern         string
	includeHeaders     bool
	includeQueryParams bool
	queryParams        map[string]bool
	next               ika.Handler
	log                *slog.Logger
}

func Factory() ika.PluginFactory {
	return &plugin{}
}

func (*plugin) Name() string {
	return "access-log"
}

func (*plugin) New(ctx context.Context, ictx ika.InjectionContext, config map[string]any) (ika.Plugin, error) {
	p := &plugin{}

	if err := pluginutil.UnmarshalCfg(config, &p.cfg); err != nil {
		return nil, err
	}

	p.queryParams = make(map[string]bool, len(p.cfg.QueryParams))
	for _, param := range p.cfg.QueryParams {
		p.queryParams[param] = true
	}

	p.ikaPattern = ictx.Route
	p.log = ictx.Logger
	p.includeHeaders = len(p.cfg.Headers) > 0
	p.includeQueryParams = len(p.cfg.QueryParams) > 0

	return p, nil
}

func (p *plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
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

func (*plugin) Teardown(context.Context) error { return nil }

func (p *plugin) makeReqAttrs(r *http.Request) []any {
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

	if p.includeQueryParams {
		query := p.getQueryVals(r)
		if len(query) == 0 {
			return requestAttrs
		}

		attrs := make([]any, 0, len(query))
		for key, values := range query {
			if len(values) == 1 {
				attrs = append(attrs, slog.String(key, values[0]))
				continue
			}
			attrs = append(attrs, slog.Any(key, values))
		}
		requestAttrs = append(requestAttrs, slog.Group("query", attrs...))
	}

	return requestAttrs
}

// getQueryVals returns the query values for the keys in p.queryParams
// https://github.com/golang/go/issues/50034
func (p *plugin) getQueryVals(r *http.Request) url.Values {
	query := r.URL.RawQuery
	m := make(url.Values, len(p.queryParams))
	var key string
	var err error
	for query != "" {
		key, query, _ = strings.Cut(query, "&")
		if key == "" {
			continue
		}

		key, value, _ := strings.Cut(key, "=")
		if !p.queryParams[key] {
			// try to unescape the key
			key, err = url.QueryUnescape(key)
			if err != nil {
				continue
			}
		}

		if !p.queryParams[key] {
			continue
		}

		m[key] = append(m[key], value)
	}
	return m
}

var (
	_ ika.OnRequestHook = &plugin{}
	_ ika.PluginFactory = &plugin{}
)
