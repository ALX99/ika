package proxy

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/plugin"
)

type Config struct {
	Transport  http.RoundTripper
	Namespace  string
	BufferPool httputil.BufferPool
}

type Proxy struct {
	rp httputil.ReverseProxy
}

type keyErr struct{}

func NewProxy(cfg Config) (*Proxy, error) {
	rp := httputil.ReverseProxy{
		BufferPool: cfg.BufferPool,
		Transport:  cfg.Transport,
		ErrorLog:   log.New(slogIOWriter{}, "httputil.ReverseProxy ", log.LstdFlags),
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			newReq := r.WithContext(context.WithValue(r.Context(), keyErr{}, err))
			*r = *newReq
		},

		Rewrite: func(rp *httputil.ProxyRequest) {
			// Restore the query even if it can't be parsed (see [httputil.ReverseProxy])
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery
		},
	}

	return &Proxy{rp: rp}, nil
}

func (p *Proxy) WithPathTrim(trim string) plugin.HandlerFunc {
	return plugin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, trim)
		r.URL.RawPath = strings.TrimPrefix(request.GetPath(r), trim)
		return p.ServeHTTP(w, r)
	})
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	p.rp.ServeHTTP(w, r)

	err := r.Context().Value(keyErr{})
	if err != nil {
		return err.(error)
	}
	return nil
}

type slogIOWriter struct{}

func (slogIOWriter) Write(p []byte) (n int, err error) {
	slog.Error(string(p))
	return len(p), nil
}
