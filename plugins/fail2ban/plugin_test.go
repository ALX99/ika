package fail2ban

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil/httperr"
	"github.com/matryer/is"
)

func TestPlugin_Setup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    map[string]any
		wantError bool
	}{
		{
			name: "valid config",
			config: map[string]any{
				"maxAttempts": uint64(3),
				"window":      "1m",
				"banDuration": "2m",
			},
			wantError: false,
		},
		{
			name: "invalid maxAttempts",
			config: map[string]any{
				"maxAttempts": uint64(0),
				"window":      "1m",
			},
			wantError: true,
		},
		{
			name: "invalid window",
			config: map[string]any{
				"maxAttempts": uint64(3),
				"window":      "0s",
			},
			wantError: true,
		},
		{
			name: "default ban duration",
			config: map[string]any{
				"maxAttempts": uint64(3),
				"window":      "1m",
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
			p := &Plugin{
				attempts: &sync.Map{},
				log:      slog.New(slog.DiscardHandler),
			}
			err := p.Setup(t.Context(), ika.InjectionContext{}, tt.config)
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

	tests := []struct {
		name           string
		maxAttempts    uint64
		window         time.Duration
		banDuration    time.Duration
		idHeader       string
		requests       []request
		wantBanned     bool
		wantStatusCode int
	}{
		{
			name:        "under max attempts",
			maxAttempts: 3,
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
			maxAttempts: 3,
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
			maxAttempts: 2,
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
			maxAttempts: 2,
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
			p := &Plugin{
				cfg: pConfig{
					MaxAttempts: tt.maxAttempts,
					Window:      tt.window,
					BanDuration: tt.banDuration,
					IDHeader:    tt.idHeader,
				},
				attempts: &sync.Map{},
				log:      slog.New(slog.DiscardHandler),
				next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					return httperr.New(http.StatusUnauthorized)
				}),
			}

			// Run the sequence of requests
			for _, req := range tt.requests {
				r := httptest.NewRequest("GET", "/", nil)
				r.RemoteAddr = req.ip
				for k, v := range req.headers {
					r.Header.Set(k, v)
				}
				w := httptest.NewRecorder()

				err := p.ServeHTTP(w, r)
				is.True(err != nil)

				var httpErr *httperr.Error
				if errors.As(err, &httpErr) {
					is.Equal(httpErr.Status(), req.wantStatus)
				}
			}

			// Verify final state
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.requests[len(tt.requests)-1].ip
			if tt.idHeader != "" {
				r.Header.Set(tt.idHeader, tt.requests[len(tt.requests)-1].headers[tt.idHeader])
			}
			w := httptest.NewRecorder()

			err := p.ServeHTTP(w, r)
			is.True(err != nil)

			var httpErr *httperr.Error
			if errors.As(err, &httpErr) {
				is.Equal(httpErr.Status(), tt.wantStatusCode)
			}
		})
	}
}

func TestPlugin_Cleanup(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	p := &Plugin{
		cfg: pConfig{
			MaxAttempts: 2,
			Window:      50 * time.Millisecond,
			BanDuration: 50 * time.Millisecond,
		},
		attempts: &sync.Map{},
		log:      slog.New(slog.DiscardHandler),
		next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return httperr.New(http.StatusUnauthorized)
		}),
	}

	// Make requests to get banned
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.1:1234"

	for i := 0; i < 2; i++ {
		err := p.ServeHTTP(httptest.NewRecorder(), r)
		is.True(err != nil)
	}

	// Verify banned
	err := p.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	var httpErr *httperr.Error
	is.True(errors.As(err, &httpErr))
	is.Equal(httpErr.Status(), http.StatusTooManyRequests)

	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)

	// Should be unbanned
	err = p.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	is.True(errors.As(err, &httpErr))
	is.Equal(httpErr.Status(), http.StatusUnauthorized)
}

type request struct {
	ip         string
	headers    map[string]string
	wantStatus int
}
