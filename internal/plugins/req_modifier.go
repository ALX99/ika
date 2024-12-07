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
	"github.com/alx99/ika/middleware"
	"github.com/alx99/ika/plugin"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

var (
	_ plugin.Plugin          = &reqModifier{}
	_ plugin.RequestModifier = &reqModifier{}
)

type RewriterFactory struct{}

func (RewriterFactory) New(context.Context) (plugin.Plugin, error) {
	return &reqModifier{}, nil
}

// reqModifier is a reqModifier that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type reqModifier struct {
	// segments is a map of segment index to their corresponding replacement
	segments []string

	// replacePattern is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replacePattern string

	host   string
	scheme string

	pathRewriteEnabled bool
	hostRewriteEnabled bool
}

func (reqModifier) Name() string {
	return "req-modifier"
}

func (reqModifier) Capabilities() []plugin.Capability {
	return []plugin.Capability{plugin.CapModifyRequests}
}

func (reqModifier) InjectionLevels() []plugin.InjectionLevel {
	return []plugin.InjectionLevel{plugin.LevelPath}
}

func (rm *reqModifier) Setup(ctx context.Context, context plugin.InjectionContext, config map[string]any) error {
	routePattern := context.PathPattern
	isNamespaced := strings.HasPrefix(context.Namespace, "/")

	var toPath string
	if _, ok := config["path"]; ok {
		toPath = config["path"].(string)
	}

	var host string
	if _, ok := config["host"]; ok {
		host = config["host"].(string)
	}

	if toPath != "" {
		rm.pathRewriteEnabled = true
		rm.setupPathRewrite(routePattern, isNamespaced, toPath)
	}

	if host != "" {
		rm.hostRewriteEnabled = true
		if err := rm.setupHostRewrite(host); err != nil {
			return err
		}
	}

	return nil
}

func (rm *reqModifier) ModifyRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
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

func (reqModifier) Teardown(context.Context) error { return nil }

func (rm *reqModifier) rewritePath(r *http.Request) error {
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
	log := slog.With(slog.String("namespace", middleware.GetMetadata(r.Context()).Namespace))
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

	log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", prevPath), slog.String("to", r.URL.RawPath))
	return nil
}

func (rm *reqModifier) rewriteHost(r *http.Request) {
	r.Host = rm.host
	r.URL.Host = rm.host
	r.URL.Scheme = rm.scheme
}

// setupPathRewrite sets up the path rewrite
func (rm *reqModifier) setupPathRewrite(routePattern string, isNamespaced bool, toPath string) {
	rm.segments = make([]string, len(strings.Split(routePattern, "/"))+1)
	s := strings.Split(routePattern, "/")

	if isNamespaced {
		// The first path segment of a namespaced route is the namespace itself
		rm.replacePattern = strings.Split(routePattern, "/")[0]
	}
	rm.replacePattern += segmentRe.ReplaceAllString(toPath, "%s")

	matches := segmentRe.FindAllStringSubmatch(toPath, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range s {
			if v == match[0] {
				if isNamespaced {
					// If a route is namespaced, the first segment is the namespace
					// which is impossible to to match with a rewritePath
					rm.segments[i+1] = match[0]
				} else {
					rm.segments[i] = match[0]
				}
			}
		}
	}
}

// setupHostRewrite sets up the host rewrite
func (rm *reqModifier) setupHostRewrite(host string) error {
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
