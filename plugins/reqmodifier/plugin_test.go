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
		name          string
		config        map[string]any
		route         string
		inputURL      string
		expectedPath  string
		expectedHost  string
		expectedQuery string
	}{
		{
			name: "simple path rewrite",
			config: map[string]any{
				"path": "/api/v2/{id}",
			},
			route:         "/users/{id}",
			inputURL:      "http://example.com/users/123",
			expectedPath:  "/api/v2/123",
			expectedQuery: "",
		},
		{
			name: "path rewrite with encoded characters",
			config: map[string]any{
				"path": "/api/v2/{id}",
			},
			route:         "/users/{id}",
			inputURL:      "http://example.com/users/test%20space",
			expectedPath:  "/api/v2/test space",
			expectedQuery: "",
		},
		{
			name: "path rewrite with wildcard",
			config: map[string]any{
				"path": "/api/{wildcard...}",
			},
			route:         "/users/{wildcard...}",
			inputURL:      "http://example.com/users/123/posts/456",
			expectedPath:  "/api/123/posts/456",
			expectedQuery: "",
		},
		{
			name: "host rewrite",
			config: map[string]any{
				"host": "https://api.example.com",
			},
			inputURL:      "http://example.com/users/123",
			expectedPath:  "/users/123",
			expectedHost:  "api.example.com",
			expectedQuery: "",
		},
		{
			name: "host rewrite with retain header",
			config: map[string]any{
				"host":             "https://api.example.com",
				"retainHostHeader": true,
			},
			inputURL:      "http://example.com/users/123",
			expectedPath:  "/users/123",
			expectedHost:  "api.example.com",
			expectedQuery: "",
		},
		{
			name: "query params are passed correctly - simple",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=1&bye=2",
			expectedPath:  "/any",
			expectedQuery: "hi=1&bye=2",
		},
		{
			name: "query params are passed correctly - encoded",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=%20hello%20world%20",
			expectedPath:  "/any",
			expectedQuery: "hi=%20hello%20world%20",
		},
		{
			name: "query params are passed correctly - multiple values",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=1&hi=2&hi=3",
			expectedPath:  "/any",
			expectedQuery: "hi=1&hi=2&hi=3",
		},
		{
			name: "wildcard rewrite with query params",
			config: map[string]any{
				"path": "/{any...}",
			},
			route:         "/httpbun/{any...}",
			inputURL:      "http://example.com/httpbun/any/a/huhh?abc=lol&x=b",
			expectedPath:  "/any/a/huhh",
			expectedQuery: "abc=lol&x=b",
		},
		{
			name: "wildcard rewrite with encoded query params",
			config: map[string]any{
				"path": "/{any...}",
			},
			route:         "/httpbun/{any...}",
			inputURL:      "http://example.com/httpbun/any/a/huhh?abc=魚&x=は",
			expectedPath:  "/any/a/huhh",
			expectedQuery: "abc=魚&x=は",
		},
		{
			name: "path rewrite with multiple segments",
			config: map[string]any{
				"path": "/any/{a1}/{a2}",
			},
			route:         "/path-rewrite/{a1}/{a2}",
			inputURL:      "http://example.com/path-rewrite/a/huhh",
			expectedPath:  "/any/a/huhh",
			expectedQuery: "",
		},
		{
			name: "retain host header",
			config: map[string]any{
				"path":             "/any",
				"host":             "http://localhost:8080",
				"retainHostHeader": true,
			},
			route:         "/retain-host",
			inputURL:      "http://example.com/retain-host",
			expectedPath:  "/any",
			expectedHost:  "localhost:8080",
			expectedQuery: "",
		},
		{
			name: "query params - single param",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=1",
			expectedPath:  "/any",
			expectedQuery: "hi=1",
		},
		{
			name: "query params - empty values",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=&bye=",
			expectedPath:  "/any",
			expectedQuery: "hi=&bye=",
		},
		{
			name: "query params - null value",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=null",
			expectedPath:  "/any",
			expectedQuery: "hi=null",
		},
		{
			name: "query params - boolean values",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=true&bye=false",
			expectedPath:  "/any",
			expectedQuery: "hi=true&bye=false",
		},
		{
			name: "query params - numeric values",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=123&bye=456.789",
			expectedPath:  "/any",
			expectedQuery: "hi=123&bye=456.789",
		},
		{
			name: "query params - multiple types",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=1&bye=true&foo=null&bar=%20space%20",
			expectedPath:  "/any",
			expectedQuery: "hi=1&bye=true&foo=null&bar=%20space%20",
		},
		{
			name: "query params - encoded path characters",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/get",
			inputURL:      "http://example.com/get?hi=hello%20world&bye=goodbye%2Fworld",
			expectedPath:  "/any",
			expectedQuery: "hi=hello%20world&bye=goodbye%2Fworld",
		},
		{
			name: "non-terminated path with trailing slash",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/not-terminated/{any}/",
			inputURL:      "http://example.com/not-terminated/hi/",
			expectedPath:  "/any",
			expectedQuery: "",
		},
		{
			name: "non-terminated path with multiple segments",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/not-terminated/{any}/",
			inputURL:      "http://example.com/not-terminated/a/b/c/",
			expectedPath:  "/any",
			expectedQuery: "",
		},
		{
			name: "non-terminated path without trailing slash",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/not-terminated/{any}/",
			inputURL:      "http://example.com/not-terminated/a/b/c/d",
			expectedPath:  "/any",
			expectedQuery: "",
		},
		{
			name: "terminated path with trailing slash",
			config: map[string]any{
				"path": "/any",
			},
			route:         "/terminated/{any}/{$}",
			inputURL:      "http://example.com/terminated/hi/",
			expectedPath:  "/any",
			expectedQuery: "",
		},
		{
			name: "wildcard path with encoded slashes",
			config: map[string]any{
				"path": "/{any...}",
			},
			route:         "/httpbun/{any...}",
			inputURL:      "http://example.com/httpbun/any/slash%2Fshould-bekept/next",
			expectedPath:  "/any/slash/should-bekept/next",
			expectedQuery: "",
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

			is.Equal(req.URL.Path, tt.expectedPath)
			if tt.expectedHost != "" {
				is.Equal(req.URL.Host, tt.expectedHost)
				if tt.config["retainHostHeader"] == true {
					is.Equal(req.Host, "example.com")
				} else {
					is.Equal(req.Host, tt.expectedHost)
				}
			}
			is.Equal(req.URL.RawQuery, tt.expectedQuery)
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
