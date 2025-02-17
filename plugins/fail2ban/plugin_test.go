package fail2ban

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil/httperr"
	"github.com/matryer/is"
)

func TestPlugin_Setup(t *testing.T) {
	t.Parallel()

	factory := Factory()

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"maxRetries":  uint64(3),
				"window":      "1m",
				"banDuration": "2m",
			},
			wantError: false,
		},
		{
			name: "invalid maxRetries",
			config: map[string]any{
				"maxRetries": uint64(0),
				"window":     "1m",
			},
			wantError: true,
		},
		{
			name: "invalid window",
			config: map[string]any{
				"maxRetries": uint64(3),
				"window":     "0s",
			},
			wantError: true,
		},
		{
			name: "default ban duration",
			config: map[string]any{
				"maxRetries": uint64(3),
				"window":     "1m",
			},
			wantError: false,
		},
		{
			name: "invalid ban duration",
			config: map[string]any{
				"banDuration": "0s",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			_, err := factory.New(t.Context(), ika.InjectionContext{
				Logger: slog.New(slog.DiscardHandler),
			}, tt.config)

			if tt.wantError {
				is.True(err != nil)
			} else {
				is.NoErr(err)
			}
		})
	}
}

func TestPlugin_ServeHTTP(t *testing.T) {
	is := is.New(t)
	factory := Factory()

	tests := []struct {
		name           string
		maxRetries     uint64
		window         time.Duration
		banDuration    time.Duration
		idHeader       string
		requests       []request
		wantBanned     bool
		wantStatusCode int
	}{
		{
			name:        "under max attempts",
			maxRetries:  3,
			window:      time.Minute,
			banDuration: time.Minute,
			requests: []request{
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
			},
			wantBanned:     false,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:        "exactly max attempts",
			maxRetries:  3,
			window:      time.Minute,
			banDuration: time.Minute,
			requests: []request{
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
			},
			wantBanned:     true,
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:        "different IPs not affecting each other",
			maxRetries:  2,
			window:      time.Minute,
			banDuration: time.Minute,
			requests: []request{
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.2:1234", wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.1:1234", wantStatus: http.StatusUnauthorized},
			},
			wantBanned:     true,
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:        "custom header identifier",
			maxRetries:  2,
			window:      time.Minute,
			banDuration: time.Minute,
			idHeader:    "X-Real-IP",
			requests: []request{
				{ip: "192.0.2.1:1234", headers: map[string]string{"X-Real-IP": "10.0.0.1"}, wantStatus: http.StatusUnauthorized},
				{ip: "192.0.2.2:1234", headers: map[string]string{"X-Real-IP": "10.0.0.1"}, wantStatus: http.StatusUnauthorized},
			},
			wantBanned:     true,
			wantStatusCode: http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p, err := factory.New(t.Context(), ika.InjectionContext{
				Logger: slog.New(slog.DiscardHandler),
			}, map[string]any{
				"maxRetries":  tt.maxRetries,
				"window":      tt.window.String(),
				"banDuration": tt.banDuration.String(),
				"idHeader":    tt.idHeader,
			})
			is.NoErr(err)

			plugin := p.(*plugin)
			plugin.next = ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
				return httperr.New(http.StatusUnauthorized)
			})

			// Run requests
			for _, req := range tt.requests {
				r := httptest.NewRequest("GET", "/", nil)
				r.RemoteAddr = req.ip
				for k, v := range req.headers {
					r.Header.Set(k, v)
				}

				err := plugin.ServeHTTP(httptest.NewRecorder(), r)
				is.True(err != nil)
				var httpErr *httperr.Error
				is.True(errors.As(err, &httpErr))
				is.Equal(httpErr.Status(), req.wantStatus)
			}

			// Verify final state
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.requests[len(tt.requests)-1].ip
			if tt.idHeader != "" {
				r.Header.Set(tt.idHeader, tt.requests[len(tt.requests)-1].headers[tt.idHeader])
			}
			err = plugin.ServeHTTP(httptest.NewRecorder(), r)
			is.True(err != nil)
			var httpErr *httperr.Error
			is.True(errors.As(err, &httpErr))
			is.Equal(httpErr.Status(), tt.wantStatusCode)
		})
	}
}

func TestPlugin_Cleanup(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	factory := Factory()

	p, err := factory.New(t.Context(), ika.InjectionContext{
		Logger: slog.New(slog.DiscardHandler),
	}, map[string]any{
		"maxRetries":  uint64(2),
		"window":      "50ms",
		"banDuration": "50ms",
	})
	is.NoErr(err)

	plugin := p.(*plugin)
	plugin.next = ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return httperr.New(http.StatusUnauthorized)
	})

	// Make requests to get banned
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.1:1234"

	for i := 0; i < 2; i++ {
		err := plugin.ServeHTTP(httptest.NewRecorder(), r)
		is.True(err != nil)
	}

	// Verify banned
	err = plugin.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	var httpErr *httperr.Error
	is.True(errors.As(err, &httpErr))
	is.Equal(httpErr.Status(), http.StatusTooManyRequests)

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Should be unbanned
	err = plugin.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	is.True(errors.As(err, &httpErr))
	is.Equal(httpErr.Status(), http.StatusUnauthorized)
}

type request struct {
	ip         string
	headers    map[string]string
	wantStatus int
}
