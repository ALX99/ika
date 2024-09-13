package main

import (
	"log"
	"os"

	"github.com/alx99/ika"
	"github.com/alx99/ika/middleware"
	"github.com/grafana/pyroscope-go"

	chimw "github.com/go-chi/chi/v5/middleware"
)

func init() {
	err := middleware.RegisterFunc("noCache", chimw.NoCache)
	if err != nil {
		panic(err)
	}
}

func main() {
	defer setupMonitoring()()
	ika.Run()
}

func setupMonitoring() func() {
	p, err := pyroscope.Start(pyroscope.Config{
		ApplicationName:   "ika-example",
		ServerAddress:     "https://profiles-prod-019.grafana.net",
		BasicAuthUser:     os.Getenv("PYROSCOPE_USER"),
		BasicAuthPassword: os.Getenv("PYROSCOPE_PASSWORD"),
	})
	if err != nil {
		log.Println("failed to start pyroscope", err)
		return func() {}
	}

	return func() {
		p.Flush(true)
	}
}
