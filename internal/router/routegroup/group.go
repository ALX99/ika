package routegroup

import (
	"cmp"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/alx99/ika/plugin"
)

// Group represents a group of routes with associated middleware.
type Group struct {
	mux         *http.ServeMux
	middlewares []func(plugin.Handler) plugin.Handler

	method       string
	path         string
	host         string
	handleCalled bool
}

// New creates a new Group.
func New(mux *http.ServeMux) *Group {
	return &Group{mux: mux}
}

// Mount creates a new group from a [http.ServeMux] pattern
func Mount(mux *http.ServeMux, pattern string) *Group {
	g := Group{mux: mux}
	g.method, g.host, g.path = decomposePattern(pattern)
	return &g
}

// Mount creates a new group with a specified base path on top of the existing bundle.
// TODO use a pattern
func (g *Group) Mount(path string) *Group {
	newGrp := g.clone()
	newGrp.path += path
	return newGrp
}

// Use adds middleware(s) to the Group.
//
// ## The middleware(s) are executed in the order added
//
// New(mux).Use(m1, m2, m3).Handle(h)
//
// is equivalent to:
//
//	m1(m2(m3(h)))
func (g *Group) Use(middlewares ...func(plugin.Handler) plugin.Handler) *Group {
	if g.handleCalled {
		panic("routergroup: tried to add middleware after HandleFunc. This is most likely a mistake.")
	}
	g.middlewares = append(g.middlewares, middlewares...)
	return g
}

// With creates a new group with the same configuration except it appends the
// given middlewares to the new group.
//
// It is exactly the same as [Group.Use] except that it does not modify the underlying group.
func (g *Group) With(middlewares ...func(plugin.Handler) plugin.Handler) *Group {
	return g.clone().Use(middlewares...)
}

// Handle adds a new route to the Group's mux, applying all middlewares to the handler.
func (g *Group) Handle(pattern string, handler plugin.Handler) *Group {
	g.HandleFunc(pattern, handler.ServeHTTP)
	return g
}

// HandleFunc registers the handler function for the given pattern to the Group's mux.
// The handler is wrapped with the Group's middlewares.
func (g *Group) HandleFunc(pattern string, handler plugin.HandlerFunc) *Group {
	g.mux.Handle(g.makePattern(pattern), g.wrapMiddleware(handler).ToHTTPHandler(nil))
	g.handleCalled = true
	return g
}

// Handler exposes the underlying [http.ServeMux.Handler] method
func (g *Group) Handler(r *http.Request) (h http.Handler, pattern string) {
	return g.mux.Handler(r)
}

// ServeHTTP implements the http.Handler interface
func (g *Group) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

func (g *Group) clone() *Group {
	newGrp := *g
	newGrp.middlewares = slices.Clone(g.middlewares)
	newGrp.handleCalled = false
	return &newGrp
}

func (g *Group) makePattern(pattern string) string {
	method, host, path := decomposePattern(pattern)

	if method != "" && g.method != "" && method != g.method {
		panic(fmt.Sprintf("routergroup: impossible route: method %s does not match the group's base method %s", method, g.method))
	}
	if host != "" && g.host != "" && host != g.host {
		panic(fmt.Sprintf("routergroup: impossible route: host %s does not match the group's base host %s", host, g.host))
	}

	// Prefer to use the method/host from the pattern
	method, host = cmp.Or(method, g.method), cmp.Or(host, g.host)

	return strings.TrimLeft(method+" "+host+g.path+path, " \t")
}

// wrapMiddleware applies the registered middlewares to a handler.
func (g *Group) wrapMiddleware(handler plugin.Handler) plugin.HandlerFunc {
	for i := range g.middlewares {
		handler = g.middlewares[len(g.middlewares)-1-i](handler)
	}
	return plugin.HandlerFunc(handler.ServeHTTP)
}

var patternRe = regexp.MustCompile(`^(?:(\S*)\s+)*\s*([^/]*)(/.*)$`)

func decomposePattern(pattern string) (method, host, path string) {
	matches := patternRe.FindStringSubmatch(pattern)
	if len(matches) < 4 {
		matches = append(matches, make([]string, 4-len(matches))...)
	}
	return matches[1], matches[2], matches[3]
}
