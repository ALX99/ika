// Package reqmodifier contains a plugin for modifying HTTP requests in the ika API Gateway.
package reqmodifier

import (
	"context"
	"errors"
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

// segmentPattern matches path segments like {id} or {wildcard...}
var segmentPattern = regexp.MustCompile(`\{([^{}]*)\}`)

type plugin struct {
	cfg pConfig

	// pathSegments maps segment positions to their names
	// Example: For /users/{id}/posts/{type}, the map would be:
	// {1: "{id}", 3: "{type}"}
	pathSegments map[int]string

	// pathTemplate is the format string for path rewriting
	// Example: For target path "/api/%s/v2/%s", segments are replaced with %s
	pathTemplate string

	// targetHostTemplate and targetScheme for host rewriting
	targetHostTemplate string
	targetScheme       string

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

	if ictx.Scope != ika.ScopeRoute {
		return nil, errors.New("plugin only usable in route scope")
	}

	// Initialize pathSegments map
	p.pathSegments = make(map[int]string)

	// Extract path segments from route pattern
	_, _, path := decomposePattern(ictx.Route)
	routeSegments := strings.Split(path, "/")[1:] // Skip empty first segment

	if p.cfg.Host != "" {
		if err := p.setupHostRewrite(p.cfg.Host, routeSegments); err != nil {
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

	if p.targetHostTemplate != "" {
		p.rewriteHost(r)
	}

	return nil
}

func (p *plugin) rewritePath(r *http.Request) error {
	path := request.GetPath(r)
	args := make([]any, 0, 8)

	// Skip leading slash and process path segments
	pos := 1
	segmentIdx := 0
	for i := 1; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			// Skip empty segments
			if i <= pos {
				pos = i + 1
				continue
			}

			segment, exists := p.pathSegments[segmentIdx]
			if exists {
				if strings.HasSuffix(segment, "...}") {
					// Handle wildcard by capturing remaining path
					args = append(args, path[pos:])
					break
				}
				args = append(args, path[pos:i])
			}
			segmentIdx++
			pos = i + 1
		}
	}

	// Format new path using collected segments
	newPath := fmt.Sprintf(p.pathTemplate, args...)

	var err error
	r.URL.RawPath = newPath
	r.URL.Path, err = url.PathUnescape(newPath)
	if err != nil {
		return err
	}

	p.log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", path), slog.String("to", r.URL.RawPath))
	return nil
}

func (p *plugin) rewriteHost(r *http.Request) {
	path := request.GetPath(r)
	args := make([]any, 0, 8)

	// Skip leading slash and process path segments
	pos := 1
	segmentIdx := 0
	for i := 1; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			// Skip empty segments
			if i <= pos {
				pos = i + 1
				continue
			}

			segment, exists := p.pathSegments[segmentIdx]
			if exists {
				if strings.HasSuffix(segment, "...}") {
					// Handle wildcard by capturing remaining path
					args = append(args, path[pos:])
					break
				}
				args = append(args, path[pos:i])
			}
			segmentIdx++
			pos = i + 1
		}
	}

	// Format new path using collected segments
	newHost := fmt.Sprintf(p.targetHostTemplate, args...)

	p.log.LogAttrs(r.Context(), slog.LevelDebug, "Host rewritten",
		slog.String("from", path), slog.String("to", newHost))

	if !p.cfg.RetainHostHeader {
		r.Host = newHost // this overrides the Host header
	}
	r.URL.Host = newHost
	r.URL.Scheme = p.targetScheme
}

func (p *plugin) setupPathRewrite(routePattern string) {
	_, _, path := decomposePattern(routePattern)
	routeSegments := strings.Split(path, "/")[1:] // Skip empty first segment

	p.pathSegments = make(map[int]string)
	p.pathTemplate = segmentPattern.ReplaceAllString(p.cfg.Path, "%s")

	// Map segment positions to their names
	for _, match := range segmentPattern.FindAllStringSubmatch(p.cfg.Path, -1) {
		if match[1] == "$" {
			continue // Skip special tokens
		}
		for i, segment := range routeSegments {
			if segment == match[0] {
				p.pathSegments[i] = match[0]
			}
		}
	}
}

func (p *plugin) setupHostRewrite(host string, routeSegments []string) error {
	scheme, host, ok := strings.Cut(host, "://")
	if !ok {
		return fmt.Errorf("invalid host URL: %s", host)
	}

	// Store the original host template for later use
	originalHost := host
	host = segmentPattern.ReplaceAllString(host, "%s")

	// Map segment positions to their names
	for _, match := range segmentPattern.FindAllStringSubmatch(originalHost, -1) {
		if match[1] == "$" {
			continue // Skip special tokens
		}

		// Find this segment in the route pattern
		for i, segment := range routeSegments {
			if segment == match[0] {
				p.pathSegments[i] = match[0]
			}
		}
	}

	p.targetHostTemplate = host
	p.targetScheme = scheme
	return nil
}

func (*plugin) Teardown(context.Context) error { return nil }

// Helper function to parse route pattern into method, host, and path components
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
