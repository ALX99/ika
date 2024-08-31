package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"

	"github.com/alx99/ika/internal/config"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

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

func (p *Proxy) Route(rewritePath string, backends []config.Backend) (http.Handler, error) {
	backend := backends[0]
	if len(backends) > 1 {
		panic("not implemented")
	}
	rw := newRewriter(rewritePath)

	rp := &httputil.ReverseProxy{
		Transport: p.transport,
		Rewrite: func(rp *httputil.ProxyRequest) {
			rp.Out.URL.Scheme = backend.Scheme
			rp.Out.URL.Host = backend.Host
			rp.Out.Host = backend.Host

			// Restore the query even if it can't be parsed
			rp.Out.URL.RawQuery = rp.In.URL.RawQuery

			if rewritePath != "" {
				rp.Out.URL.Path = rw.rewrite(rp.In)
			}
		},
	}

	return rp, nil
}

type rewriter struct {
	// toPattern is the path which the request will be rewritten
	toPattern string
	// segments is a map of segment names to their corresponding replacement
	segments map[string]string
}

func newRewriter(to string) rewriter {
	rw := rewriter{segments: make(map[string]string), toPattern: to}

	matches := segmentRe.FindAllStringSubmatch(to, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}
		rw.segments[strings.TrimSuffix(match[1], "...")] = match[0]
	}
	return rw
}

func (rw rewriter) rewrite(r *http.Request) string {
	args := make([]string, 0, len(rw.segments)*2)
	for segment, replace := range rw.segments {
		// If the segment is suffixed with '...}', then the segment is a wildcard
		// and so we don't want to escape the value.
		if strings.HasSuffix(replace, "...}") {
			args = append(args, replace, r.PathValue(segment))
		} else {
			// Otherwise, escape the value to ensure that values such
			// as '/' are safely encoded in the new path
			args = append(args, replace, url.PathEscape(r.PathValue(segment)))
		}
	}

	return strings.NewReplacer(args...).Replace(rw.toPattern)
}
