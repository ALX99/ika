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

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
		check     func(*is.I, *Plugin)
	}{
		{
			name:      "valid config with defaults",
			config:    map[string]any{},
			wantError: false,
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Header, "X-Request-ID")
				is.Equal(p.cfg.Variant, vXID)
				is.True(*p.cfg.Override)
				is.Equal(p.cfg.Append, false)
				is.True(*p.cfg.Expose)

				// Test ID generation
				id, err := p.genID()
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
			check: func(is *is.I, p *Plugin) {
				is.Equal(*p.cfg.Expose, false)
				is.Equal(p.cfg.Header, "X-Request-ID") // Should use default header
			},
		},
		{
			name: "valid config with UUIDv4",
			config: map[string]any{
				"variant": "UUIDv4",
				"header":  "X-Correlation-ID",
			},
			wantError: false,
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Header, "X-Correlation-ID")
				is.Equal(p.cfg.Variant, vUUIDv4)

				// Test ID generation
				id, err := p.genID()
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
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Variant, vUUIDv7)

				// Test ID generation
				id, err := p.genID()
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
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Variant, vKSUID)

				// Test ID generation
				id, err := p.genID()
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
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Append, true)
				is.Equal(*p.cfg.Override, false)

				// Test that ID generation still works
				id, err := p.genID()
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
			check: func(is *is.I, p *Plugin) {
				is.Equal(p.cfg.Header, "X-Request-ID") // Should use default header
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

			p := &Plugin{}
			err := p.Setup(t.Context(), ika.InjectionContext{
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

	tests := []struct {
		name           string
		config         map[string]any
		setupHeader    string
		wantHeader     bool
		wantRespHeader bool
	}{
		{
			name:           "default config adds header and response",
			config:         map[string]any{},
			wantHeader:     true,
			wantRespHeader: true,
		},
		{
			name: "respects custom header name",
			config: map[string]any{
				"header": "X-Trace-ID",
			},
			wantHeader:     true,
			wantRespHeader: true,
		},
		{
			name: "expose disabled",
			config: map[string]any{
				"expose": false,
			},
			wantHeader:     true,
			wantRespHeader: false,
		},
		{
			name: "expose with existing header",
			config: map[string]any{
				"expose":   true,
				"override": false,
			},
			setupHeader:    "test-id",
			wantHeader:     true,
			wantRespHeader: true,
		},
		{
			name: "expose with override",
			config: map[string]any{
				"expose":   true,
				"override": true,
			},
			setupHeader:    "test-id",
			wantHeader:     true,
			wantRespHeader: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			p := &Plugin{}
			err := p.Setup(t.Context(), ika.InjectionContext{
				Logger: slog.New(slog.DiscardHandler),
			}, tt.config)
			is.NoErr(err)

			req := httptest.NewRequest("GET", "/", nil)
			if tt.setupHeader != "" {
				headerName := "X-Request-ID"
				if h, ok := tt.config["header"].(string); ok {
					headerName = h
				}
				req.Header.Set(headerName, tt.setupHeader)
			}

			w := httptest.NewRecorder()
			err = p.Handler(ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				return nil
			})).ServeHTTP(w, req)
			is.NoErr(err)

			headerName := "X-Request-ID"
			if h, ok := tt.config["header"].(string); ok {
				headerName = h
			}

			if tt.wantHeader {
				is.True(req.Header.Get(headerName) != "")
			}

			if tt.wantRespHeader {
				is.True(w.Header().Get(headerName) != "")
				// Response header should match request header
				is.Equal(w.Header().Get(headerName), req.Header.Get(headerName))
			} else {
				is.Equal(w.Header().Get(headerName), "")
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
