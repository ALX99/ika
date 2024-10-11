package proxy

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/middleware"
)

type Config struct {
	Transport      http.RoundTripper
	RoutePattern   string
	IsNamespaced   bool
	Namespace      string
	RewritePattern config.Nullable[string]
	Backends       []config.Backend
}

func NewProxy(cfg Config) *httputil.ReverseProxy {
	backend := cfg.Backends[0]
	if len(cfg.Backends) > 1 {
		panic("not implemented")
	}

	var rw pathRewriter = newIndexRewriter(cfg.RoutePattern, cfg.IsNamespaced, cfg.RewritePattern.V)

	rp := &httputil.ReverseProxy{
		BufferPool: bufPool,
		Transport:  cfg.Transport,
		ErrorLog:   log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host
			// Restore the query even if it can't be parsed (see [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if !cfg.RewritePattern.Set() {
				// If no rewrite path is set, and the route is namespaced, we will strip the namespace from the path
				if cfg.IsNamespaced && cfg.Namespace != "root" {
					setPath(rp, strings.TrimPrefix(request.GetPath(rp.In), "/"+cfg.Namespace))
				}
				return
			}
			setPath(rp, rw.rewrite(rp.In))
		},
	}

	return rp
}

// setPath sets the path on the outgoing request
func setPath(rp *httputil.ProxyRequest, path string) {
	log := slog.With(slog.String("namespace", middleware.GetMetadata(rp.In.Context()).Namespace))
	var err error
	prevPath := request.GetPath(rp.Out)
	rp.Out.URL.RawPath = path
	rp.Out.URL.Path, err = url.PathUnescape(path)
	if err != nil {
		log.LogAttrs(rp.In.Context(), slog.LevelError, "impossible error made possible",
			slog.String("err", err.Error()))
	} else {
		// remove query params from the path
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
