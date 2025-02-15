package accesslog

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
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
				is.Equal(len(plugin.cfg.Headers), 1)
				is.Equal(plugin.cfg.Headers[0], "X-Request-ID")
				is.Equal(plugin.cfg.RemoteAddr, false)
				is.Equal(plugin.includeHeaders, true)
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
		name          string
		config        map[string]any
		request       *http.Request
		wantLogFields map[string]any
	}{
		{
			name:   "basic request",
			config: map[string]any{},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Pattern = "/test"
				return req
			}(),
			wantLogFields: map[string]any{
				"request.method":        "GET",
				"request.path":          "/test",
				"request.host":          "example.com",
				"request.pattern":       "/test",
				"response.duration":     float64(0),
				"response.status":       float64(200),
				"response.bytesWritten": float64(0),
				"ika.pattern":           "/test",
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
			wantLogFields: map[string]any{
				"request.headers.X-Request-ID": "123",
				"request.headers.User-Agent":   "test-agent",
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
			wantLogFields: map[string]any{
				"request.remoteAddr": "127.0.0.1:1234",
			},
		},
		{
			name: "with selected query parameters",
			config: map[string]any{
				"queryParams": []string{"foo", "multi"},
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?foo=bar&baz=qux&multi=1&multi=2", nil)
				req.Pattern = "/test"
				return req
			}(),
			wantLogFields: map[string]any{
				"request.query.foo":   "bar",
				"request.query.multi": []any{"1", "2"},
			},
		},
		{
			name: "with query parameters not present",
			config: map[string]any{
				"queryParams": []string{"notfound"},
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?foo=bar", nil)
				req.Pattern = "/test"
				return req
			}(),
			wantLogFields: map[string]any{
				"request.method":        "GET",
				"request.path":          "/test",
				"request.host":          "example.com",
				"request.pattern":       "/test",
				"response.duration":     float64(0),
				"response.status":       float64(200),
				"response.bytesWritten": float64(0),
				"ika.pattern":           "/test",
			},
		},
		{
			name: "with encoded query parameter",
			config: map[string]any{
				"queryParams": []string{"あ"},
			},
			request: func() *http.Request {
				req := httptest.NewRequest("GET", "/test?%e3%81%82=hi%20there", nil)
				req.Pattern = "/test"
				return req
			}(),
			wantLogFields: map[string]any{
				"request.query.あ": "hi%20there",
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
				w.WriteHeader(200)
				return nil
			})

			err = plugin.ServeHTTP(httptest.NewRecorder(), tt.request)
			is.NoErr(err)

			// Split log entries (there might be multiple JSON objects)
			logEntries := make([]map[string]any, 0)
			for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
				if line == "" {
					continue
				}
				var entry map[string]any
				err = json.Unmarshal([]byte(line), &entry)
				is.NoErr(err)
				logEntries = append(logEntries, entry)
			}

			// The last entry should be the access log
			accessLog := logEntries[len(logEntries)-1]
			for field, want := range tt.wantLogFields {
				got := getField(accessLog, field)
				is.Equal(got, want) // field value should match expected
			}
		})
	}
}

func TestPlugin_getQueryVals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		queryParams []string
		rawQuery    string
		want        url.Values
	}{
		{
			name:        "empty query",
			queryParams: []string{"foo"},
			rawQuery:    "",
			want:        url.Values{},
		},
		{
			name:        "simple key-value",
			queryParams: []string{"foo"},
			rawQuery:    "foo=bar",
			want:        url.Values{"foo": []string{"bar"}},
		},
		{
			name:        "multiple key-value pairs",
			queryParams: []string{"foo", "baz"},
			rawQuery:    "foo=bar&baz=qux",
			want:        url.Values{"foo": []string{"bar"}, "baz": []string{"qux"}},
		},
		{
			name:        "multiple values for same key",
			queryParams: []string{"foo"},
			rawQuery:    "foo=bar&foo=baz",
			want:        url.Values{"foo": []string{"bar", "baz"}},
		},
		{
			name:        "empty value",
			queryParams: []string{"foo"},
			rawQuery:    "foo=",
			want:        url.Values{"foo": []string{""}},
		},
		{
			name:        "no value",
			queryParams: []string{"foo"},
			rawQuery:    "foo",
			want:        url.Values{"foo": []string{""}},
		},
		{
			name:        "encoded query",
			queryParams: []string{"あ"},
			rawQuery:    "%e3%81%82=hi%20there",
			want:        url.Values{"あ": []string{"hi%20there"}},
		},
		{
			name:        "ignore non-configured params",
			queryParams: []string{"foo"},
			rawQuery:    "foo=bar&baz=qux",
			want:        url.Values{"foo": []string{"bar"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			var buf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			p := &plugin{
				log:         logger,
				queryParams: make(map[string]bool),
			}
			for _, param := range tt.queryParams {
				p.queryParams[param] = true
			}

			r := httptest.NewRequest("GET", "/?"+tt.rawQuery, nil)

			got := p.getQueryVals(r)
			is.Equal(got, tt.want)
		})
	}
}

// getField gets a nested field value from a map using dot notation
func getField(m map[string]any, field string) any {
	var current any = m
	parts := strings.Split(field, ".")
	for i, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil
			}
			if i == len(parts)-1 {
				return current
			}
		case map[string][]any:
			if val, ok := v[part]; ok && i == len(parts)-1 {
				return val
			}
			return nil
		default:
			return nil
		}
	}
	return current
}
