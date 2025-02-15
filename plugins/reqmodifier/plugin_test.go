package reqmodifier

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func TestPlugin_Setup(t *testing.T) {
	t.Parallel()

	factory := Factory()

	tests := []struct {
		name      string
		config    map[string]any
		route     string
		wantError bool
		check     func(*is.I, ika.Plugin)
	}{
		{
			name:      "empty config",
			config:    map[string]any{},
			wantError: true,
		},
		{
			name: "valid path rewrite",
			config: map[string]any{
				"path": "/new/{id}",
			},
			route:     "/old/{id}",
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(plugin.cfg.Path, "/new/{id}")
			},
		},
		{
			name: "valid host rewrite",
			config: map[string]any{
				"host": "https://example.com",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(plugin.cfg.Host, "https://example.com")
				is.Equal(plugin.host, "example.com")
				is.Equal(plugin.scheme, "https")
			},
		},
		{
			name: "valid path and host rewrite",
			config: map[string]any{
				"path": "/new/{id}",
				"host": "https://example.com",
			},
			route:     "/old/{id}",
			wantError: false,
		},
		{
			name: "path rewrite without route pattern",
			config: map[string]any{
				"path": "/new/{id}",
			},
			wantError: true,
		},
		{
			name: "invalid host URL",
			config: map[string]any{
				"host": "://invalid",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			p, err := factory.New(t.Context(), ika.InjectionContext{
				Route:  tt.route,
				Logger: slog.New(slog.DiscardHandler),
			}, tt.config)

			if tt.wantError {
				is.True(err != nil)
				return
			}

			is.NoErr(err)
			if tt.check != nil {
				tt.check(is, p)
			}
		})
	}
}

func TestPlugin_ModifyRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    map[string]any
		route     string
		request   *http.Request
		wantPath  string
		wantHost  string
		wantError bool
	}{
		{
			name: "path rewrite with single segment",
			config: map[string]any{
				"path": "/new/{id}",
			},
			route: "/old/{id}",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "/old/123", nil)
				r.Pattern = "/old/{id}"
				return r
			}(),
			wantPath: "/new/123",
		},
		{
			name: "path rewrite with wildcard",
			config: map[string]any{
				"path": "/new/{wildcard...}",
			},
			route: "/old/{wildcard...}",
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "/old/a/b/c", nil)
				r.Pattern = "/old/{wildcard...}"
				return r
			}(),
			wantPath: "/new/a/b/c",
		},
		{
			name: "host rewrite",
			config: map[string]any{
				"host": "https://example.com",
			},
			request:  httptest.NewRequest("GET", "/test", nil),
			wantHost: "example.com",
		},
		{
			name: "host rewrite with retain header",
			config: map[string]any{
				"host":             "https://example.com",
				"retainHostHeader": true,
			},
			request: func() *http.Request {
				r := httptest.NewRequest("GET", "/test", nil)
				r.Host = "original.com"
				return r
			}(),
			wantHost: "original.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			p, err := Factory().New(t.Context(), ika.InjectionContext{
				Route:  tt.route,
				Logger: slog.New(slog.DiscardHandler),
			}, tt.config)
			is.NoErr(err)

			err = p.(*plugin).ModifyRequest(tt.request)
			if tt.wantError {
				is.True(err != nil)
				return
			}

			is.NoErr(err)
			if tt.wantPath != "" {
				is.Equal(tt.request.URL.Path, tt.wantPath)
			}
			if tt.wantHost != "" {
				is.Equal(tt.request.Host, tt.wantHost)
			}
		})
	}
}

func TestDecomposePattern(t *testing.T) {
	tests := []struct {
		name       string
		pattern    string
		wantMethod string
		wantHost   string
		wantPath   string
	}{
		{
			name:     "path only",
			pattern:  "/users/{id}",
			wantPath: "/users/{id}",
		},
		{
			name:       "method and path",
			pattern:    "GET /users/{id}",
			wantMethod: "GET",
			wantPath:   "/users/{id}",
		},
		{
			name:     "host and path",
			pattern:  "example.com/users/{id}",
			wantHost: "example.com",
			wantPath: "/users/{id}",
		},
		{
			name:       "method, host and path",
			pattern:    "GET example.com/users/{id}",
			wantMethod: "GET",
			wantHost:   "example.com",
			wantPath:   "/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			method, host, path := decomposePattern(tt.pattern)
			is.Equal(method, tt.wantMethod)
			is.Equal(host, tt.wantHost)
			is.Equal(path, tt.wantPath)
		})
	}
}

func BenchmarkRewritePath(b *testing.B) {
	is := is.New(b)
	config := map[string]any{
		"path": "/new/{path}",
	}
	iCtx := ika.InjectionContext{
		Route:  "/old/{path}",
		Logger: slog.New(slog.DiscardHandler),
	}
	p, err := Factory().New(b.Context(), iCtx, config)
	if err != nil {
		b.Fatal(err)
	}
	rm := p.(*plugin)

	rm.setupPathRewrite(iCtx.Route)

	req, _ := http.NewRequest("GET", "http://example.com/old/test", nil)

	for b.Loop() {
		req.URL.Path = "/old/test"
		err := rm.rewritePath(req)
		is.NoErr(err)

		if req.URL.Path != "/new/test" {
			b.Fatalf("unexpected path: %s", req.URL.Path)
		}
	}
}
