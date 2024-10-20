package proxy

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/alx99/ika/internal/request"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

type pathRewriter interface {
	rewrite(r *http.Request) (rawPath string)
}

// indexRewriter is a rewriter that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type indexRewriter struct {
	// segments is a map of segment index to their corresponding replacement
	segments []string

	// replacePattern is in the format of /example/%s/path
	// where %s should be replaced with the corresponding segment
	replacePattern string
	pool           *sync.Pool
}

func newIndexRewriter(routePattern string, isNamespaced bool, toPattern string) indexRewriter {
	rw := indexRewriter{
		segments: make([]string, len(strings.Split(routePattern, "/"))+1),
		pool:     &sync.Pool{New: func() any { b := make([]any, 0, 10); return &b }},
	}
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
	return rw
}

func (ar indexRewriter) rewrite(r *http.Request) string {
	reqPath := strings.Split(request.GetPath(r), "/")

	args := make([]any, 0, 10)

	for segmentIndex, repl := range ar.segments {
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
	return fmt.Sprintf(ar.replacePattern, args...)
}

func isWildcard(segment string) bool {
	return strings.HasSuffix(segment, "...}")
}
