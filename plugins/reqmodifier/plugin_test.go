package reqmodifier

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func TestPlugin_New(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]any
		route       string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid path rewrite",
			config: map[string]any{
				"path": "/new/{id}",
			},
			route: "/old/{id}",
		},
		{
			name: "valid host rewrite",
			config: map[string]any{
				"host": "https://example.com",
			},
		},
		{
			name: "valid path and host rewrite",
			config: map[string]any{
				"path": "/new/{id}",
				"host": "https://example.com",
			},
			route: "/old/{id}",
		},
		{
			name: "path rewrite without route pattern",
			config: map[string]any{
				"path": "/new/{id}",
			},
			wantErr:     true,
			errContains: "path pattern is required",
		},
		{
			name: "invalid host URL",
			config: map[string]any{
				"host": "://invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			iCtx := ika.InjectionContext{
				Route:  tt.route,
				Logger: slog.New(slog.DiscardHandler),
			}

			p, err := (&Plugin{}).New(context.Background(), iCtx, tt.config)
			if tt.wantErr {
				is.True(err != nil)
				if tt.errContains != "" {
					is.True(err.Error() == tt.errContains)
				}
				return
			}

			is.NoErr(err)
			is.True(p != nil)
		})
	}
}

func TestPlugin_ModifyRequest(t *testing.T) {
	tests := []struct {
		name           string
		config         map[string]any
		route          string
		inputURL       string
		expectedPath   string
		expectedHost   string
		expectedScheme string
	}{
		{
			name: "simple path rewrite",
			config: map[string]any{
				"path": "/api/v2/{id}",
			},
			route:        "/users/{id}",
			inputURL:     "http://example.com/users/123",
			expectedPath: "/api/v2/123",
		},
		{
			name: "path rewrite with encoded characters",
			config: map[string]any{
				"path": "/api/v2/{id}",
			},
			route:        "/users/{id}",
			inputURL:     "http://example.com/users/test%20space",
			expectedPath: "/api/v2/test space",
		},
		{
			name: "path rewrite with wildcard",
			config: map[string]any{
				"path": "/api/{wildcard...}",
			},
			route:        "/users/{wildcard...}",
			inputURL:     "http://example.com/users/123/posts/456",
			expectedPath: "/api/123/posts/456",
		},
		{
			name: "host rewrite",
			config: map[string]any{
				"host": "https://api.example.com",
			},
			inputURL:       "http://example.com/users/123",
			expectedHost:   "api.example.com",
			expectedScheme: "https",
		},
		{
			name: "host rewrite with retain header",
			config: map[string]any{
				"host":             "https://api.example.com",
				"retainHostHeader": true,
			},
			inputURL:       "http://example.com/users/123",
			expectedHost:   "api.example.com",
			expectedScheme: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := is.New(t)
			iCtx := ika.InjectionContext{
				Route:  tt.route,
				Logger: slog.New(slog.DiscardHandler),
			}

			p, err := (&Plugin{}).New(context.Background(), iCtx, tt.config)
			is.NoErr(err)

			req, err := http.NewRequest(http.MethodGet, tt.inputURL, nil)
			is.NoErr(err)
			if tt.route != "" {
				req.Pattern = tt.route
			}

			rm := p.(*Plugin)
			err = rm.ModifyRequest(req)
			is.NoErr(err)

			if tt.expectedPath != "" {
				is.Equal(req.URL.Path, tt.expectedPath)
			}
			if tt.expectedHost != "" {
				is.Equal(req.URL.Host, tt.expectedHost)
				if tt.config["retainHostHeader"] == true {
					is.Equal(req.Host, "example.com")
				} else {
					is.Equal(req.Host, tt.expectedHost)
				}
			}
			if tt.expectedScheme != "" {
				is.Equal(req.URL.Scheme, tt.expectedScheme)
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
	p, err := (&Plugin{}).New(b.Context(), iCtx, config)
	if err != nil {
		b.Fatal(err)
	}
	rm := p.(*Plugin)

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
