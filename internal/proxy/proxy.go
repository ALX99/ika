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
	rp        *httputil.ReverseProxy
}

type Config struct {
	Transport      http.RoundTripper
	RoutePattern   string
	IsNamespaced   bool
	Namespace      string
	RewritePattern config.Nullable[string]
	Backends       []config.Backend
}

func NewProxy(cfg Config) *Proxy {
	backend := cfg.Backends[0]
	if len(cfg.Backends) > 1 {
		panic("not implemented")
	}

	p := &Proxy{transport: cfg.Transport}

	var rw pathRewriter = newIndexRewriter(cfg.RoutePattern, cfg.IsNamespaced, cfg.RewritePattern.V)

	p.rp = &httputil.ReverseProxy{
		Transport: p.transport,
		ErrorLog:  log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host
			// Restore the query even if it can't be parsed (read [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if !cfg.RewritePattern.Set() {
				// If no rewrite path is set, and the route is namespaced, we will strip the namespace from the path
				if cfg.IsNamespaced && cfg.Namespace != "root" {
					setPath(rp, strings.TrimPrefix(rp.In.URL.EscapedPath(), "/"+cfg.Namespace))
				}
				return
			}
			setPath(rp, rw.rewrite(rp.In))
		},
	}

	return p
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.rp.ServeHTTP(w, r)
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
	} else {
		// remove qurey params from the path
		rp.Out.URL.Path = strings.SplitN(rp.Out.URL.Path, "?", 2)[0]
	}

	log.LogAttrs(rp.In.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", prevPath), slog.String("to", rp.Out.URL.RawPath))
}

type slogIOWriter struct{}

func (slogIOWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
