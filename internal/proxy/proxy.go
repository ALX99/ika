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
)

type Config struct {
	Transport  http.RoundTripper
	Namespace  string
	Backends   []config.Backend
	BufferPool httputil.BufferPool
}

func NewProxy(cfg Config) (*httputil.ReverseProxy, error) {
	backend := cfg.Backends[0]
	if len(cfg.Backends) > 1 {
		panic("not implemented")
	}

	u, err := url.Parse(backend.Host)
	if err != nil {
		return nil, err
	}

	rp := &httputil.ReverseProxy{
		BufferPool: cfg.BufferPool,
		Transport:  cfg.Transport,
		ErrorLog:   log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),

		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = u.Scheme
			rp.Out.URL.Host = u.Host
			rp.Out.Host = u.Host
			// Restore the query even if it can't be parsed (see [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			// todo unacceptable hack to trim the namespace from the path
			rp.Out.URL.Path = strings.TrimPrefix(rp.In.URL.Path, cfg.Namespace)
			rp.Out.URL.RawPath = strings.TrimPrefix(request.GetPath(rp.In), cfg.Namespace)
		},
	}

	return rp, nil
}

type slogIOWriter struct{}

func (slogIOWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
