package accesslog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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
		wantError bool
		check     func(*is.I, ika.Plugin)
	}{
		{
			name:      "empty config",
			config:    map[string]any{},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(len(plugin.cfg.Headers), 0)
				is.Equal(plugin.cfg.RemoteAddr, false)
			},
		},
		{
			name: "with headers",
			config: map[string]any{
				"headers": []string{"X-Request-ID", "User-Agent"},
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(len(plugin.cfg.Headers), 2)
				is.Equal(plugin.cfg.Headers[0], "X-Request-ID")
				is.Equal(plugin.cfg.Headers[1], "User-Agent")
				is.Equal(plugin.includeHeaders, true)
			},
		},
		{
			name: "with remote addr",
			config: map[string]any{
				"remoteAddr": true,
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(plugin.cfg.RemoteAddr, true)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			p, err := factory.New(t.Context(), ika.InjectionContext{
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

func TestPlugin_ServeHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		config         map[string]any
		request        *http.Request
		responseStatus int
		wantLogFields  []string
	}{
		{
			name:   "basic request",
			config: map[string]any{},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Pattern = "/test"
				return req
			}(),
			responseStatus: http.StatusOK,
			wantLogFields: []string{
				"request.method",
				"request.path",
				"request.host",
				"request.pattern",
				"response.duration",
				"response.status",
				"response.bytesWritten",
				"ika.pattern",
			},
		},
		{
			name: "with headers",
			config: map[string]any{
				"headers": []string{"X-Request-ID", "User-Agent"},
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Pattern = "/test"
				req.Header.Set("X-Request-ID", "123")
				req.Header.Set("User-Agent", "test-agent")
				return req
			}(),
			responseStatus: http.StatusOK,
			wantLogFields: []string{
				"request.headers.X-Request-ID",
				"request.headers.User-Agent",
			},
		},
		{
			name: "with remote addr",
			config: map[string]any{
				"remoteAddr": true,
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Pattern = "/test"
				req.RemoteAddr = "127.0.0.1:1234"
				return req
			}(),
			responseStatus: http.StatusOK,
			wantLogFields: []string{
				"request.remoteAddr",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			factory := Factory()
			p, err := factory.New(context.Background(), ika.InjectionContext{
				Logger: logger,
				Route:  tt.request.Pattern,
			}, tt.config)
			is.NoErr(err)

			plugin := p.(*plugin)
			plugin.next = ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(tt.responseStatus)
				return nil
			})

			err = plugin.ServeHTTP(httptest.NewRecorder(), tt.request)
			is.NoErr(err)

			// Verify log output
			var logEntry map[string]any
			err = json.Unmarshal(buf.Bytes(), &logEntry)
			is.NoErr(err)

			for _, field := range tt.wantLogFields {
				is.True(hasField(logEntry, field)) // field should exist in log output
			}
		})
	}
}

// hasField checks if a nested field exists in a map using dot notation
func hasField(m map[string]any, field string) bool {
	var current any = m
	for _, part := range strings.Split(field, ".") {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return false
			}
		default:
			return false
		}
	}
	return true
}
