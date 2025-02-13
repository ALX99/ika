package fail2ban

import (
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

func TestPlugin_ServeHTTP(t *testing.T) {
	is := is.New(t)

	p := &Plugin{
		cfg: pConfig{
			MaxAttempts: 3,
			Window:      time.Minute,
			BanDuration: time.Minute,
		},
		attempts: &sync.Map{},
		log:      slog.New(slog.DiscardHandler),
		next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return httperr.New(http.StatusUnauthorized)
		}),
	}

	// Make requests from same IP
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "192.0.2.1:1234"

	for range 3 {
		err := p.ServeHTTP(httptest.NewRecorder(), r)
		is.True(err != nil)
		is.Equal(err.(*httperr.Error).Status(), http.StatusUnauthorized)
	}

	// Fourth attempt should be banned
	err := p.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	is.Equal(err.(*httperr.Error).Status(), http.StatusTooManyRequests)
}

func TestPlugin_CustomIdentifierHeader(t *testing.T) {
	is := is.New(t)

	p := &Plugin{
		cfg: pConfig{
			MaxAttempts: 3,
			Window:      time.Minute,
			BanDuration: time.Minute,
			IDHeader:    "X-Custom-ID",
		},
		attempts: &sync.Map{},
		log:      slog.New(slog.DiscardHandler),
		next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
			return httperr.New(http.StatusUnauthorized)
		}),
	}

	// Make request with custom identifier header
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Custom-ID", "user123")
	r.RemoteAddr = "192.0.2.1:1234" // Should be ignored since we're using custom header

	// Should track custom identifier
	for range 3 {
		err := p.ServeHTTP(httptest.NewRecorder(), r)
		is.True(err != nil)
	}

	// Should be banned by custom identifier
	err := p.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	is.Equal(err.(*httperr.Error).Status(), http.StatusTooManyRequests)

	// Different identifier should not be banned
	r.Header.Set("X-Custom-ID", "user456")
	err = p.ServeHTTP(httptest.NewRecorder(), r)
	is.True(err != nil)
	is.Equal(err.(*httperr.Error).Status(), http.StatusUnauthorized)
}
