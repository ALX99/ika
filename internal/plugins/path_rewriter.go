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
	_ plugin.Plugin          = &rewriter{}
	_ plugin.RequestModifier = rewriter{}
)

type RewriterFactory struct{}

func (RewriterFactory) New(context.Context) (plugin.Plugin, error) {
	return &rewriter{}, nil
}

// rewriter is a rewriter that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type rewriter struct {
	// segments is a map of segment index to their corresponding replacement
	segments []string

	// replacePattern is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replacePattern string
}

func (rewriter) Name() string {
	return "path-rewriter"
}

func (rewriter) Capabilities() []plugin.Capability {
	return []plugin.Capability{plugin.CapModifyRequests}
}

func (rw *rewriter) Setup(ctx context.Context, context plugin.Context, config map[string]any) error {
	routePattern := context.PathPattern
	isNamespaced := strings.HasPrefix(context.Namespace, "/")
	toPattern := config["to"].(string)

	rw.segments = make([]string, len(strings.Split(routePattern, "/"))+1)
	s := strings.Split(routePattern, "/")

	if isNamespaced {
		// The first path segment of a namespaced route is the namespace itself
		rw.replacePattern = strings.Split(routePattern, "/")[0]
	}
	rw.replacePattern += segmentRe.ReplaceAllString(toPattern, "%s")

	matches := segmentRe.FindAllStringSubmatch(toPattern, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range s {
			if v == match[0] {
				if isNamespaced {
					// If a route is namespaced, the first segment is the namespace
					// which is impossible to to match with a rewritePath
					rw.segments[i+1] = match[0]
				} else {
					rw.segments[i] = match[0]
				}
			}
		}
	}

	return nil
}

func (rw rewriter) ModifyRequest(ctx context.Context, r *http.Request) (*http.Request, error) {
	reqPath := strings.Split(request.GetPath(r), "/")

	args := make([]any, 0, 10)

	for segmentIndex, repl := range rw.segments {
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
	path := fmt.Sprintf(rw.replacePattern, args...)
	var err error

	r.URL.RawPath = path
	r.URL.Path, err = url.PathUnescape(path)
	if err != nil {
		return nil, err
	}
	// remove query params from the path
	r.URL.Path = strings.SplitN(r.URL.Path, "?", 2)[0]

	log.LogAttrs(r.Context(), slog.LevelDebug, "Path rewritten",
		slog.String("from", prevPath), slog.String("to", r.URL.RawPath))
	return r, nil
}

func (rewriter) Teardown(context.Context) error { return nil }

func isWildcard(segment string) bool {
	return strings.HasSuffix(segment, "...}")
}
