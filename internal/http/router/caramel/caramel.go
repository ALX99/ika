package caramel

import (
	"cmp"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/alx99/ika/plugin"
)

type Caramel struct {
	mux         *http.ServeMux
	middlewares []func(plugin.Handler) plugin.Handler

	method       string
	path         string
	host         string
	handleCalled bool
}

// Wrap creates a new Caramel instance that will register patterns
// to the provided mux
func Wrap(mux *http.ServeMux) *Caramel {
	return &Caramel{mux: mux}
}

// Mount takes the [http.ServeMux] pattern and creates a new Caramel instance from it.
func (c *Caramel) Mount(pattern string) *Caramel {
	method, host, path := decomposePattern(pattern)
	if path == "" {
		host = pattern // special case, treat the pattern as a host
	}
	c.assertMountable(method, host)

	newGrp := c.clone()
	newGrp.method = cmp.Or(method, c.method)
	newGrp.host = cmp.Or(host, c.host)
	newGrp.path += path

	return newGrp
}

// Use adds middleware(t) that will automatically be applied when [Caramel.HandleFunc] or [Caramel.Handle] is called.
func (c *Caramel) Use(middlewares ...func(plugin.Handler) plugin.Handler) *Caramel {
	if c.handleCalled {
		panic("caramel: tried to add middleware after HandleFunc. This is most likely a mistake.")
	}
	c.middlewares = append(c.middlewares, middlewares...)
	return c
}

// With creates a new Caramel instance with the same configuration as the original, but with additional middlewares.
func (c *Caramel) With(middlewares ...func(plugin.Handler) plugin.Handler) *Caramel {
	return c.clone().Use(middlewares...)
}

// Handle is the equivilant of [http.ServeMux.Handle]
func (c *Caramel) Handle(pattern string, handler plugin.Handler) *Caramel {
	c.HandleFunc(pattern, handler.ServeHTTP)
	return c
}

// HandleFunc is the equivilant of [http.ServeMux.HandleFunc]
func (c *Caramel) HandleFunc(pattern string, handler plugin.HandlerFunc) *Caramel {
	c.mux.Handle(c.makePattern(pattern), c.wrapMiddleware(handler).ToHTTPHandler(nil))
	c.handleCalled = true
	return c
}

func (c *Caramel) clone() *Caramel {
	newGrp := *c
	newGrp.middlewares = slices.Clone(c.middlewares)
	newGrp.handleCalled = false
	return &newGrp
}

func (c *Caramel) makePattern(pattern string) string {
	method, host, path := decomposePattern(pattern)
	c.assertMountable(method, host)
	method, host = cmp.Or(method, c.method), cmp.Or(host, c.host)
	return strings.TrimLeft(method+" "+host+c.path+path, " \t")
}

func (c *Caramel) assertMountable(method, host string) {
	if method != "" && c.method != "" && method != c.method {
		panic(fmt.Sprintf("caramel: impossible route: method %s does not match the base method %s", method, c.method))
	}
	if host != "" && c.host != "" && host != c.host {
		panic(fmt.Sprintf("caramel: impossible route: host %s does not match the base host %s", host, c.host))
	}
}

// wrapMiddleware applies the registered middlewares to a handler.
func (c *Caramel) wrapMiddleware(handler plugin.Handler) plugin.HandlerFunc {
	for i := range c.middlewares {
		handler = c.middlewares[len(c.middlewares)-1-i](handler)
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
