package proxy

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// regular expression to match segments in the rewrite path
var segmentRe = regexp.MustCompile(`\{([^{}]*)\}`)

type pathRewriter interface {
	rewrite(r *http.Request) (rawPath string)
}

// indexRewriter is a rewriter that 100% accurately rewrite the request path.
// This includes totally preserving the original path even if some parts have been encoded.
type indexRewriter struct {
	// toPattern is the path which the request will be rewritten
	toPattern string
	// segments is a map of segment index to their corresponding replacement
	segments map[int]string
}

func newIndexRewriter(fromPattern, toPattern string) indexRewriter {
	rw := indexRewriter{segments: make(map[int]string), toPattern: toPattern}
	s := strings.Split(fromPattern, "/")

	matches := segmentRe.FindAllStringSubmatch(toPattern, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}

		for i, v := range s {
			if v == match[0] {
				rw.segments[i] = match[0]
			}
		}
	}
	return rw
}

func (ar indexRewriter) rewrite(r *http.Request) string {
	args := make([]string, 0, len(ar.segments)*2)
	s := strings.Split(r.URL.EscapedPath(), "/")

	for segmentIndex, replace := range ar.segments {
		if segmentIndex >= len(s) {
			continue
		}

		if isWildcard(replace) {
			args = append(args, replace, strings.Join(s[segmentIndex:], "/"))
			continue
		}

		args = append(args, replace, s[segmentIndex])
	}

	return strings.NewReplacer(args...).Replace(ar.toPattern)
}

// valueRewriter is a rewriter will rewrite the request path using
// request.PathValue() to get the value of the segment.
// However, this might not 100% preserve encoded parts of the original path.
// Specifically for wildcard segments, the segment will be decoded
// before being replaced in the new path.
type valueRewriter struct {
	// toPattern is the path which the request will be rewritten
	toPattern string
	// segments is a map of segment names to their corresponding replacement
	segments map[string]string
}

func newValueRewriter(toPattern string) valueRewriter {
	rw := valueRewriter{segments: make(map[string]string), toPattern: toPattern}

	matches := segmentRe.FindAllStringSubmatch(toPattern, -1)
	for _, match := range matches {
		if match[1] == "$" {
			continue // special token, not a segment
		}
		rw.segments[strings.TrimSuffix(match[1], "...")] = match[0]
	}
	return rw
}

func (rw valueRewriter) rewrite(r *http.Request) string {
	args := make([]string, 0, len(rw.segments)*2)
	for segName, replace := range rw.segments {
		if isWildcard(replace) {
			val := r.PathValue(segName)

			// Dilemma! Impossible to tell if the slash was originally '/' or '%2F'
			// The only safe fallback is to keep the unescaped value
			if strings.Contains(val, "/") {
				args = append(args, replace, r.PathValue(segName))
				continue
			}
		}

		// Otherwise, escape the value to ensure that values such
		args = append(args, replace, url.PathEscape(r.PathValue(segName)))
	}

	return strings.NewReplacer(args...).Replace(rw.toPattern)
}

func isWildcard(segment string) bool {
	return strings.HasSuffix(segment, "...}")
}
