package proxy

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/alx99/ika/internal/config"
)

type Proxy struct {
	transport http.RoundTripper
}

func NewProxy(transport http.RoundTripper) *Proxy {
	return &Proxy{transport: transport}
}

func (p *Proxy) GetHandler(routePattern, rewritePattern string, backends []config.Backend) (http.Handler, error) {
	backend := backends[0]
	if len(backends) > 1 {
		panic("not implemented")
	}
	var rw pathRewriter = newIndexRewriter(routePattern, rewritePattern)

	var err error
	rp := &httputil.ReverseProxy{
		Transport: p.transport,
		ErrorLog:  log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host
			// Restore the query even if it can't be parsed (read [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if rewritePattern == "" {
				return
			}

			prevPath := rp.Out.URL.EscapedPath()
			rp.Out.URL.RawPath = rw.rewrite(rp.In)
			rp.Out.URL.Path, err = url.PathUnescape(rp.Out.URL.RawPath)
			if err != nil {
				slog.LogAttrs(rp.In.Context(), slog.LevelError, "impossible error made possible",
					slog.String("err", err.Error()))
			}

			slog.LogAttrs(rp.In.Context(), slog.LevelDebug, "Path rewritten",
				slog.String("from", prevPath), slog.String("to", rp.Out.URL.RawPath))
		},
	}

	return rp, nil
}

type slogIOWriter struct{}

func (slogIOWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
