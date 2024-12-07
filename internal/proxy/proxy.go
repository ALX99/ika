package proxy

import (
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/alx99/ika/internal/request"
)

type Config struct {
	Transport  http.RoundTripper
	Namespace  string
	BufferPool httputil.BufferPool
}

func NewProxy(cfg Config) (*httputil.ReverseProxy, error) {
	rp := &httputil.ReverseProxy{
		BufferPool: cfg.BufferPool,
		Transport:  cfg.Transport,
		ErrorLog:   log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),

		Rewrite: func(rp *httputil.ProxyRequest) {
			// rp.Out.URL.Scheme = rp.In.URL.Scheme
			// rp.Out.URL.Host = rp.In.URL.Host
			// rp.Out.Host = rp.In.Host
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
