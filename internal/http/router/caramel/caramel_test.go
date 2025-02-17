package caramel

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matryer/is"
)

func TestCaramel_patterns(t *testing.T) {
	t.Parallel()

	t.Run("longest path wins", func(t *testing.T) {
		t.Parallel()
		is := is.New(t)

		mux := http.NewServeMux()
		c := Wrap(mux).Mount("/api")

		c.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		c.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		})

		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/abcde", nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusNotFound)

		req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/users", nil)
		rr = httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		is.Equal(rr.Code, http.StatusOK)
	})
}

func TestCaramel_makePattern(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	t.Run("should return a pattern with the method and path", func(t *testing.T) {
		t.Parallel()
		c := &Caramel{path: "/api"}
		pattern := c.makePattern("/users")
		is.Equal(pattern, "/api/users")
	})

	t.Run("should handle nested groups", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		c1 := Wrap(mux).Mount("/api")
		c2 := c1.Mount("/v1")
		pattern := c2.makePattern("/users")
		is.Equal(pattern, "/api/v1/users")
	})

	t.Run("should handle nested groups with methods", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		c1 := Wrap(mux).Mount("/api")
		c2 := c1.Mount("/v1")
		pattern := c2.makePattern("GET /users")
		is.Equal(pattern, "GET /api/v1/users")
	})

	t.Run("should handle root path in nested groups", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		c1 := Wrap(mux).Mount("/api")
		c2 := c1.Mount("/v1")
		pattern := c2.makePattern("/")
		is.Equal(pattern, "/api/v1/")
	})

	t.Run("should panic if method does not match group's base method", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not get one")
			}
		}()
		mux := http.NewServeMux()
		c := Wrap(mux).Mount("POST /api")
		c.makePattern("GET /users")
	})

	t.Run("should panic if host does not match group's base host", func(t *testing.T) {
		t.Parallel()
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic but did not get one")
			}
		}()
		mux := http.NewServeMux()
		c := Wrap(mux).Mount("example.com/api")
		c.makePattern("another.com/users")
	})

	t.Run("should handle patterns with host", func(t *testing.T) {
		t.Parallel()
		c := &Caramel{path: "/api"}
		pattern := c.makePattern("example.com/users")
		is.Equal(pattern, "example.com/api/users")
	})

	t.Run("should handle patterns with method and host", func(t *testing.T) {
		t.Parallel()
		c := &Caramel{path: "/api"}
		pattern := c.makePattern("GET example.com/users")
		is.Equal(pattern, "GET example.com/api/users")
	})

	t.Run("should handle nested groups with host", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		c1 := Wrap(mux).Mount("example.com/api")
		c2 := c1.Mount("/v1")
		pattern := c2.makePattern("/users")
		is.Equal(pattern, "example.com/api/v1/users")
	})

	t.Run("should handle nested groups with method and host", func(t *testing.T) {
		t.Parallel()
		mux := http.NewServeMux()
		c1 := Wrap(mux).Mount("example.com/api")
		c2 := c1.Mount("/v1")
		pattern := c2.makePattern("GET /users")
		is.Equal(pattern, "GET example.com/api/v1/users")
	})
}

func TestDecomposePattern(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			is := is.New(t)
			method, host, path := decomposePattern(tt.pattern)
			is.Equal(method, tt.expected[0])
			is.Equal(host, tt.expected[1])
			is.Equal(path, tt.expected[2])
		})
	}
}
