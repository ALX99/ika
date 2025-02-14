// Package plugins contains built-in plugins for the ika API Gateway.
package plugins

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
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

var _ ika.RequestModifier = &ReqModifier{}

// ReqModifier is a ReqModifier that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type ReqModifier struct {
	// A map of segment index (in the route pattern)
	// to the segment name
	//
	// For example, if the route pattern is /example/{id}/path/{wildcard...}
	// segments will be {1: "{id}", 3: "{wildcard...}"}
	segments map[int]string

	// replaceFormat is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replaceFormat string

	// settings
	host               string
	scheme             string
	toPath             string
	pathRewriteEnabled bool
	hostRewriteEnabled bool
	retainHostHeader   bool

	log  *slog.Logger
	once sync.Once
}

func (*ReqModifier) New(_ context.Context, _ ika.InjectionContext) (ika.Plugin, error) {
	return &ReqModifier{}, nil
}

func (*ReqModifier) Name() string {
	return "basic-modifier"
}

func (rm *ReqModifier) Setup(ctx context.Context, ictx ika.InjectionContext, config map[string]any) error {
	routePattern := ictx.Route

	var toPath string
	if _, ok := config["path"]; ok {
		toPath = config["path"].(string)
	}

	var host string
	if _, ok := config["host"]; ok {
		host = config["host"].(string)
	}

	if _, ok := config["retainHostHeader"]; ok {
		rm.retainHostHeader = config["retainHostHeader"].(bool)
	}

	if toPath != "" {
		if routePattern == "" {
			return fmt.Errorf("path pattern is required")
		}
		rm.pathRewriteEnabled = true
		rm.toPath = toPath
	}

	if host != "" {
		rm.hostRewriteEnabled = true
		if err := rm.setupHostRewrite(host); err != nil {
			return err
		}
	}

	rm.log = ictx.Logger
	return nil
}

func (rm *ReqModifier) ModifyRequest(r *http.Request) error {
	if rm.pathRewriteEnabled {
		rm.once.Do(func() { rm.setupPathRewrite(r.Pattern) })
		if err := rm.rewritePath(r); err != nil {
			return err
		}
	}

	if rm.hostRewriteEnabled {
		rm.rewriteHost(r)
	}

	return nil
}

func (rm *ReqModifier) rewritePath(r *http.Request) error {
	var err error
	path := request.GetPath(r)
	args := make([]any, 0, 64)
	// first element is always empty due to leading slash
	splitPath := strings.Split(path, "/")[1:]
	reqPathLen := len(splitPath)

	for i, segment := range splitPath[:reqPathLen] {
		repl, ok := rm.segments[i]
		if !ok {
			continue
		}

		if isWildcard(repl) {
			args = append(args, strings.Join(splitPath[i:], "/"))
			break // bail, wildcard must always be the last segment
		}
		args = append(args, segment)
	}

	newPath := fmt.Sprintf(rm.replaceFormat, args...)

	r.URL.RawPath = newPath
	r.URL.Path, err = url.PathUnescape(newPath)
	if err != nil {
		return err
	}

	rm.log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", path), slog.String("to", r.URL.RawPath))
	return nil
}
func (*ReqModifier) Teardown(context.Context) error { return nil }

func (rm *ReqModifier) rewriteHost(r *http.Request) {
	if !rm.retainHostHeader {
		r.Host = rm.host // this overrides the Host header
	}
	r.URL.Host = rm.host
	r.URL.Scheme = rm.scheme
}

// setupPathRewrite sets up the path rewrite
func (rm *ReqModifier) setupPathRewrite(routePattern string) {
	_, _, path := decomposePattern(routePattern)
	// first element is always empty due to leading slash
	routeSplit := strings.Split(path, "/")[1:]

	rm.segments = make(map[int]string)
	rm.replaceFormat = segmentRe.ReplaceAllString(rm.toPath, "%s")

	matches := segmentRe.FindAllStringSubmatch(rm.toPath, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range routeSplit {
			if v == match[0] {
				rm.segments[i] = match[0]
			}
		}
	}
}

// setupHostRewrite sets up the host rewrite
func (rm *ReqModifier) setupHostRewrite(host string) error {
	u, err := url.Parse(host)
	if err != nil {
		return err
	}

	rm.host = u.Host
	rm.scheme = u.Scheme
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
