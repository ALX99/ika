package middlewares

import (
	"context"
	"net/http"

	"github.com/alx99/ika/middleware"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func init() {
	err := middleware.RegisterProvider("noCache", &NoCache{})
	if err != nil {
		panic(err)
	}
}

type NoCache struct{}

func (NoCache) GetMiddleware(_ context.Context, _ map[string]any) (middleware.Middleware, error) {
	return func(next http.Handler) http.Handler {
		return chimw.NoCache(next)
	}, nil
}

func (*NoCache) Teardown(_ context.Context) error {
	return nil
}
