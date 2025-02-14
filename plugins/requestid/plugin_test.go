package requestid

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func TestPlugin_ModifyRequest(t *testing.T) {
	t.Parallel()
	genID := func() (string, error) { return "request-id", nil }
	next := ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error { return nil })
	tests := []struct {
		name string
		p    Plugin
		// Named input parameters for target function.
		r              *http.Request
		wantHeader     http.Header
		wantRespHeader http.Header
	}{
		{
			name: "no override header",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{false}[0],
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"test"}},
			wantRespHeader: http.Header{"X-Request-Id": {"test"}},
		},
		{
			name: "override header",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{true}[0],
					Variant:  vUUIDv4,
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"request-id"}},
			wantRespHeader: http.Header{"X-Request-Id": {"request-id"}},
		},
		{
			name: "append header",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Append:   true,
					Variant:  vUUIDv4,
					Override: &[]bool{false}[0],
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"test", "request-id"}},
			wantRespHeader: http.Header{"X-Request-Id": {"test"}},
		},
		{
			name: "expose disabled",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{false}[0],
					Expose:   &[]bool{false}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader: http.Header{"X-Request-Id": {"test"}},
		},
		{
			name: "expose header with no existing ID",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{false}[0],
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"request-id"}},
			wantRespHeader: http.Header{"X-Request-Id": {"request-id"}},
		},
		{
			name: "expose header with existing ID",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{false}[0],
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"test"}},
			wantRespHeader: http.Header{"X-Request-Id": {"test"}},
		},
		{
			name: "expose header with override",
			p: Plugin{
				cfg: pConfig{
					Header:   "X-Request-Id",
					Override: &[]bool{true}[0],
					Expose:   &[]bool{true}[0],
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader:     http.Header{"X-Request-Id": {"request-id"}},
			wantRespHeader: http.Header{"X-Request-Id": {"request-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			w := httptest.NewRecorder()
			err := tt.p.ServeHTTP(w, tt.r)
			is.NoErr(err)

			if tt.wantHeader != nil {
				is.Equal(tt.wantHeader.Get("X-Request-Id"), tt.r.Header.Get("X-Request-Id"))
			}

			if tt.wantRespHeader != nil {
				is.Equal(tt.wantRespHeader.Get("X-Request-Id"), w.Header().Get("X-Request-Id"))
			}
		})
	}
}

func TestPlugin_Setup(t *testing.T) {
	t.Parallel()

	factory := &Plugin{}

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
		check     func(*is.I, ika.Plugin)
	}{
		{
			name:      "valid config with defaults",
			config:    map[string]any{},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Header, "X-Request-ID")
				is.Equal(plugin.cfg.Variant, vXID)
				is.True(*plugin.cfg.Override)
				is.Equal(plugin.cfg.Append, false)
				is.True(*plugin.cfg.Expose)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) > 0) // XID should generate a non-empty string
			},
		},
		{
			name: "valid config with expose disabled",
			config: map[string]any{
				"expose": false,
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(*plugin.cfg.Expose, false)
				is.Equal(plugin.cfg.Header, "X-Request-ID") // Should use default header
			},
		},
		{
			name: "valid config with UUIDv4",
			config: map[string]any{
				"variant": "UUIDv4",
				"header":  "X-Correlation-ID",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Header, "X-Correlation-ID")
				is.Equal(plugin.cfg.Variant, vUUIDv4)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) == 36) // UUIDv4 should be 36 chars
			},
		},
		{
			name: "valid config with UUIDv7",
			config: map[string]any{
				"variant": "UUIDv7",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Variant, vUUIDv7)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) == 36) // UUIDv7 should be 36 chars
			},
		},
		{
			name: "valid config with KSUID",
			config: map[string]any{
				"variant": "KSUID",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Variant, vKSUID)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) > 0) // KSUID should generate a non-empty string
			},
		},
		{
			name: "valid config with append",
			config: map[string]any{
				"append":   true,
				"override": false,
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Append, true)
				is.Equal(*plugin.cfg.Override, false)

				// Test that ID generation still works
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) > 0)
			},
		},
		{
			name: "invalid variant",
			config: map[string]any{
				"variant": "invalid",
			},
			wantError: true,
		},
		{
			name: "missing header uses default",
			config: map[string]any{
				"header": nil,
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Header, "X-Request-ID") // Should use default header
			},
		},
		{
			name: "conflicting append and override",
			config: map[string]any{
				"append":   true,
				"override": true,
			},
			wantError: true,
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

func TestPlugin_Integration(t *testing.T) {
	t.Parallel()

	factory := &Plugin{}

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
		check     func(*is.I, ika.Plugin)
	}{
		{
			name: "valid config with UUIDv4",
			config: map[string]any{
				"variant": "UUIDv4",
				"header":  "X-Correlation-ID",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Header, "X-Correlation-ID")
				is.Equal(plugin.cfg.Variant, vUUIDv4)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) == 36) // UUIDv4 should be 36 chars
			},
		},
		{
			name: "valid config with UUIDv7",
			config: map[string]any{
				"variant": "UUIDv7",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Variant, vUUIDv7)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) == 36) // UUIDv7 should be 36 chars
			},
		},
		{
			name: "valid config with KSUID",
			config: map[string]any{
				"variant": "KSUID",
			},
			wantError: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*Plugin)
				is.Equal(plugin.cfg.Variant, vKSUID)

				// Test ID generation
				id, err := plugin.genID()
				is.NoErr(err)
				is.True(len(id) > 0) // KSUID should generate a non-empty string
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

func BenchmarkIDGeneration(b *testing.B) {
	tests := []struct {
		name    string
		variant string
	}{
		{"UUIDv4", vUUIDv4},
		{"UUIDv7", vUUIDv7},
		{"KSUID", vKSUID},
		{"XID", vXID},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			genFn, err := makeRandFun(tt.variant)
			if err != nil {
				b.Fatal(err)
			}

			for b.Loop() {
				_, err := genFn()
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
