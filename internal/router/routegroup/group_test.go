package routegroup

import (
	"net/http"
	"testing"

	"github.com/matryer/is"
)

func TestGroup_makePattern(t *testing.T) {
	is := is.New(t)

	t.Run("should return a pattern with the method and path", func(t *testing.T) {
		g := &Group{path: "/api"}
		pattern := g.makePattern("/users")
		is.Equal(pattern, "/api/users")
	})

	t.Run("should handle nested groups", func(t *testing.T) {
		mux := http.NewServeMux()
		g1 := Mount(mux, "/api")
		g2 := g1.Mount("/v1")
		pattern := g2.makePattern("/users")
		is.Equal(pattern, "/api/v1/users")
	})

	t.Run("should handle nested groups with methods", func(t *testing.T) {
		mux := http.NewServeMux()
		g1 := Mount(mux, "/api")
		g2 := g1.Mount("/v1")
		pattern := g2.makePattern("GET /users")
		is.Equal(pattern, "GET /api/v1/users")
	})

	t.Run("should handle root path in nested groups", func(t *testing.T) {
		mux := http.NewServeMux()
		g1 := Mount(mux, "/api")
		g2 := g1.Mount("/v1")
		pattern := g2.makePattern("/")
		is.Equal(pattern, "/api/v1/")
	})

	t.Run("should panic if method does not match group's base method", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not get one")
			}
		}()
		mux := http.NewServeMux()
		g := Mount(mux, "POST /api")
		g.makePattern("GET /users")
	})

	t.Run("should panic if host does not match group's base host", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not get one")
			}
		}()
		mux := http.NewServeMux()
		g := Mount(mux, "example.com/api")
		g.makePattern("another.com/users")
	})

	t.Run("should handle patterns with host", func(t *testing.T) {
		g := &Group{path: "/api"}
		pattern := g.makePattern("example.com/users")
		is.Equal(pattern, "example.com/api/users")
	})

	t.Run("should handle patterns with method and host", func(t *testing.T) {
		g := &Group{path: "/api"}
		pattern := g.makePattern("GET example.com/users")
		is.Equal(pattern, "GET example.com/api/users")
	})

	t.Run("should handle nested groups with host", func(t *testing.T) {
		mux := http.NewServeMux()
		g1 := Mount(mux, "example.com/api")
		g2 := g1.Mount("/v1")
		pattern := g2.makePattern("/users")
		is.Equal(pattern, "example.com/api/v1/users")
	})

	t.Run("should handle nested groups with method and host", func(t *testing.T) {
		mux := http.NewServeMux()
		g1 := Mount(mux, "example.com/api")
		g2 := g1.Mount("/v1")
		pattern := g2.makePattern("GET /users")
		is.Equal(pattern, "GET example.com/api/v1/users")
	})
}

func TestDecomposePattern(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected [3]string
	}{
		{
			"should decompose pattern with method, host, and path",
			"GET example.com/users",
			[3]string{"GET", "example.com", "/users"},
		},
		{
			"should decompose pattern with host and path",
			"example.com/users",
			[3]string{"", "example.com", "/users"},
		},
		{
			"should decompose pattern with method and path",
			"GET /users",
			[3]string{"GET", "", "/users"},
		},
		{
			"should decompose pattern with only path",
			"/users",
			[3]string{"", "", "/users"},
		},
		{
			"should handle empty pattern",
			"",
			[3]string{"", "", ""},
		},
		{
			"should handle only host",
			"example.com/",
			[3]string{"", "example.com", "/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			method, host, path := decomposePattern(tt.pattern)
			is.Equal(method, tt.expected[0])
			is.Equal(host, tt.expected[1])
			is.Equal(path, tt.expected[2])
		})
	}
}
