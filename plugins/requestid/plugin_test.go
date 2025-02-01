package requestid

import (
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
		r          *http.Request
		wantHeader http.Header
	}{
		{
			name: "no override header",
			p: Plugin{
				cfg: config{
					Header: "X-Request-Id",
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
			name: "override header",
			p: Plugin{
				cfg: config{
					Header:   "X-Request-Id",
					Override: true,
					Variant:  vUUIDv4,
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader: http.Header{"X-Request-Id": {"request-id"}},
		},
		{
			name: "append header",
			p: Plugin{
				cfg: config{
					Header:  "X-Request-Id",
					Append:  true,
					Variant: vUUIDv4,
				},
				genID: genID,
				next:  next,
			},
			r: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Add("X-Request-Id", "test")
				return req
			}(),
			wantHeader: http.Header{"X-Request-Id": {"test", "request-id"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			err := tt.p.ServeHTTP(httptest.NewRecorder(), tt.r)
			is.NoErr(err)

			if tt.wantHeader != nil {
				is.Equal(tt.wantHeader.Get("X-Request-ID"), tt.r.Header.Get("X-Request-ID"))
			}
		})
	}
}
