package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/alx99/ika/internal/config"
)

type Proxy struct {
	transport *http.Transport
}

func NewProxy(transportCfg config.Transport) *Proxy {
	return &Proxy{
		transport: &http.Transport{
			DisableKeepAlives:      transportCfg.DisableKeepAlives,
			DisableCompression:     transportCfg.DisableCompression,
			MaxIdleConns:           transportCfg.MaxIdleConns,
			MaxIdleConnsPerHost:    transportCfg.MaxIdleConnsPerHost,
			MaxConnsPerHost:        transportCfg.MaxConnsPerHost,
			IdleConnTimeout:        transportCfg.IdleConnTimeout,
			ResponseHeaderTimeout:  transportCfg.ResponseHeaderTimeout,
			ExpectContinueTimeout:  transportCfg.ExpectContinueTimeout,
			MaxResponseHeaderBytes: transportCfg.MaxResponseHeaderBytes,
			WriteBufferSize:        transportCfg.WriteBufferSize,
			ReadBufferSize:         transportCfg.ReadBufferSize,
		},
	}
}

func (p *Proxy) Route(backends []config.Backend) (http.Handler, error) {
	backend := backends[0]
	if len(backends) > 1 {
		panic("not implemented")
	}

	rp := &httputil.ReverseProxy{
		Transport: p.transport,
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host

			// Restore the query even if it can't be parsed
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if backend.RewritePath != "" {
				panic("not implemented")
			}
		},
	}

	return rp, nil
}
