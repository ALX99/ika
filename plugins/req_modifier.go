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

	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/plugin"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

var (
	_ plugin.Plugin          = &ReqModifier{}
	_ plugin.RequestModifier = &ReqModifier{}
)

// ReqModifier is a ReqModifier that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type ReqModifier struct {
	// segments is a map of segment index to their corresponding replacement
	segments []string

	// replacePattern is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replacePattern string

	host   string
	scheme string

	pathRewriteEnabled bool
	hostRewriteEnabled bool
	retainHostHeader   bool

	log *slog.Logger
}

func (ReqModifier) New(_ context.Context, _ plugin.InjectionContext) (plugin.Plugin, error) {
	return &ReqModifier{}, nil
}

func (ReqModifier) Name() string {
	return "basic-modifier"
}

func (rm *ReqModifier) Setup(ctx context.Context, iCtx plugin.InjectionContext, config map[string]any) error {
	routePattern := iCtx.PathPattern

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
		rm.setupPathRewrite(routePattern, toPath)
	}

	if host != "" {
		rm.hostRewriteEnabled = true
		if err := rm.setupHostRewrite(host); err != nil {
			return err
		}
	}

	rm.log = iCtx.Logger
	return nil
}

func (rm *ReqModifier) ModifyRequest(r *http.Request) (*http.Request, error) {
	if rm.pathRewriteEnabled {
		if err := rm.rewritePath(r); err != nil {
			return nil, err
		}
	}

	if rm.hostRewriteEnabled {
		rm.rewriteHost(r)
	}

	return r, nil
}

func (ReqModifier) Teardown(context.Context) error { return nil }

func (rm *ReqModifier) rewritePath(r *http.Request) error {
	reqPath := strings.Split(request.GetPath(r), "/")

	args := make([]any, 0, 10)

	for segmentIndex, repl := range rm.segments {
		if repl == "" {
			continue // skip if no replacement
		}
		for i, v := range reqPath {
			if i == segmentIndex {
				args = append(args, v)
			}
			if isWildcard(repl) {
				args = append(args, strings.Join(reqPath[segmentIndex:], "/"))
				goto done // bail, wildcard must always be the last segment
			}
		}
	}

done:
	prevPath := request.GetPath(r)
	path := fmt.Sprintf(rm.replacePattern, args...)
	var err error

	r.URL.RawPath = path
	r.URL.Path, err = url.PathUnescape(path)
	if err != nil {
		return err
	}
	// remove query params from the path
	r.URL.Path = strings.SplitN(r.URL.Path, "?", 2)[0]

	rm.log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", prevPath), slog.String("to", r.URL.RawPath))
	return nil
}

func (rm *ReqModifier) rewriteHost(r *http.Request) {
	if !rm.retainHostHeader {
		r.Host = rm.host // this overrides the Host header
	}
	r.URL.Host = rm.host
	r.URL.Scheme = rm.scheme
}

// setupPathRewrite sets up the path rewrite
func (rm *ReqModifier) setupPathRewrite(routePattern string, toPath string) {
	rm.segments = make([]string, len(strings.Split(routePattern, "/"))+1)
	s := strings.Split(routePattern, "/")

	rm.replacePattern += segmentRe.ReplaceAllString(toPath, "%s")

	matches := segmentRe.FindAllStringSubmatch(toPath, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range s {
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
