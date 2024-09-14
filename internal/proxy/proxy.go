package proxy

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/middleware"
)

type Proxy struct {
	transport http.RoundTripper
}

func NewProxy(transport http.RoundTripper) *Proxy {
	return &Proxy{transport: transport}
}

func (p *Proxy) GetHandler(routePattern string, isNamespaced bool, namespace string, rewritePattern config.Nullable[string], backends []config.Backend) (http.Handler, error) {
	backend := backends[0]
	if len(backends) > 1 {
		panic("not implemented")
	}
	var rw pathRewriter = newIndexRewriter(routePattern, isNamespaced, rewritePattern.V)

	rp := &httputil.ReverseProxy{
		Transport: p.transport,
		ErrorLog:  log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host
			// Restore the query even if it can't be parsed (read [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if !rewritePattern.Set() {
				// If no rewrite path is set, and the route is namespaced, we will strip the namespace from the path
				if isNamespaced && namespace != "root" {
					setPath(rp, strings.TrimPrefix(rp.In.URL.EscapedPath(), "/"+namespace))
				}
				return
			}
			setPath(rp, rw.rewrite(rp.In))
		},
	}

	return rp, nil
}

// setPath sets the path on the outgoing request
func setPath(rp *httputil.ProxyRequest, rawPath string) {
	log := slog.With(slog.String("namespace", middleware.GetNamespace(rp.In.Context())))
	var err error
	prevPath := rp.Out.URL.EscapedPath()
	rp.Out.URL.RawPath = rawPath
	rp.Out.URL.Path, err = url.PathUnescape(rp.Out.URL.RawPath)
	if err != nil {
		log.LogAttrs(rp.In.Context(), slog.LevelError, "impossible error made possible",
			slog.String("err", err.Error()))
	}
	log.LogAttrs(rp.In.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", prevPath), slog.String("to", rp.Out.URL.RawPath))
}

type slogIOWriter struct{}

func (slogIOWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
