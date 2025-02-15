// Package reqmodifier contains a plugin for modifying HTTP requests in the ika API Gateway.
package reqmodifier

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/alx99/ika"
	"github.com/alx99/ika/internal/http/request"
	"github.com/alx99/ika/pluginutil"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

type plugin struct {
	cfg pConfig

	// A map of segment index (in the route pattern)
	// to the segment name
	//
	// For example, if the route pattern is /example/{id}/path/{wildcard...}
	// segments will be {1: "{id}", 3: "{wildcard...}"}
	segments map[int]string

	// replaceFormat is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replaceFormat string

	// host and scheme for host rewriting
	host   string
	scheme string

	log  *slog.Logger
	once sync.Once
}

func Factory() ika.PluginFactory {
	return &plugin{}
}

func (*plugin) Name() string {
	return "req-modifier"
}

func (*plugin) New(ctx context.Context, ictx ika.InjectionContext, config map[string]any) (ika.Plugin, error) {
	p := &plugin{
		log: ictx.Logger,
	}

	if err := pluginutil.UnmarshalCfg(config, &p.cfg); err != nil {
		return nil, err
	}

	if p.cfg.Path != "" {
		if ictx.Route == "" {
			return nil, fmt.Errorf("path pattern is required")
		}
	}

	if p.cfg.Host != "" {
		if err := p.setupHostRewrite(p.cfg.Host); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *plugin) ModifyRequest(r *http.Request) error {
	if p.cfg.Path != "" {
		p.once.Do(func() { p.setupPathRewrite(r.Pattern) })
		if err := p.rewritePath(r); err != nil {
			return err
		}
	}

	if p.cfg.Host != "" {
		p.rewriteHost(r)
	}

	return nil
}

func (p *plugin) rewritePath(r *http.Request) error {
	var err error
	path := request.GetPath(r)
	args := make([]any, 0, 64)
	// first element is always empty due to leading slash
	splitPath := strings.Split(path, "/")[1:]
	reqPathLen := len(splitPath)

	for i, segment := range splitPath[:reqPathLen] {
		repl, ok := p.segments[i]
		if !ok {
			continue
		}

		if isWildcard(repl) {
			args = append(args, strings.Join(splitPath[i:], "/"))
			break // bail, wildcard must always be the last segment
		}
		args = append(args, segment)
	}

	newPath := fmt.Sprintf(p.replaceFormat, args...)

	r.URL.RawPath = newPath
	r.URL.Path, err = url.PathUnescape(newPath)
	if err != nil {
		return err
	}

	p.log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", path), slog.String("to", r.URL.RawPath))
	return nil
}

func (*plugin) Teardown(context.Context) error { return nil }

func (p *plugin) rewriteHost(r *http.Request) {
	if !p.cfg.RetainHostHeader {
		r.Host = p.host // this overrides the Host header
	}
	r.URL.Host = p.host
	r.URL.Scheme = p.scheme
}

// setupPathRewrite sets up the path rewrite
func (p *plugin) setupPathRewrite(routePattern string) {
	_, _, path := decomposePattern(routePattern)
	// first element is always empty due to leading slash
	routeSplit := strings.Split(path, "/")[1:]

	p.segments = make(map[int]string)
	p.replaceFormat = segmentRe.ReplaceAllString(p.cfg.Path, "%s")

	matches := segmentRe.FindAllStringSubmatch(p.cfg.Path, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range routeSplit {
			if v == match[0] {
				p.segments[i] = match[0]
			}
		}
	}
}

// setupHostRewrite sets up the host rewrite
func (p *plugin) setupHostRewrite(host string) error {
	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	p.host = u.Host
	p.scheme = u.Scheme
	return nil
}

func isWildcard(segment string) bool {
	return strings.HasSuffix(segment, "...}")
}

var patternRe = regexp.MustCompile(`^(?:(\S*)\s+)*\s*([^/]*)(/.*)$`)

func decomposePattern(pattern string) (method, host, path string) {
	matches := patternRe.FindStringSubmatch(pattern)
	if len(matches) < 4 {
		matches = append(matches, make([]string, 4-len(matches))...)
	}
	return matches[1], matches[2], matches[3]
}

var (
	_ ika.RequestModifier = &plugin{}
	_ ika.PluginFactory   = &plugin{}
)
