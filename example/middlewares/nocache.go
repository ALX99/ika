package middlewares

import (
	"github.com/alx99/ika/middleware"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func init() {
	err := middleware.RegisterFunc("noCache", chimw.NoCache)
	if err != nil {
		panic(err)
	}
}
