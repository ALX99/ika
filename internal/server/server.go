package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/alx99/ika/internal/config"
)

// Server represents an HTTP server.
type Server struct {
	servers []http.Server
}

// NewServer creates a new server with the given handler and configuration.
func NewServer(handler http.Handler, config []config.Server) *Server {
	var servers []http.Server
	for _, c := range config {
		servers = append(servers, http.Server{
			Handler:                      handler,
			Addr:                         c.Addr,
			DisableGeneralOptionsHandler: c.DisableGeneralOptionsHandler.V,
			ReadTimeout:                  c.ReadTimeout.V,
			ReadHeaderTimeout:            c.ReadHeaderTimeout.V,
			WriteTimeout:                 c.WriteTimeout.V,
			IdleTimeout:                  c.IdleTimeout.V,
			MaxHeaderBytes:               c.MaxHeaderBytes.V,
		})
	}

	return &Server{servers: servers}
}

// ListenAndServe starts the server and listens for incoming connections.
func (s *Server) ListenAndServe() error {
	var errs error
	var mutex sync.Mutex
	wg := sync.WaitGroup{}
	for i := range s.servers {
		wg.Add(1)
		go func() {
			wg.Done()
			err := s.servers[i].ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				mutex.Lock()
				errs = errors.Join(err, s.servers[i].ListenAndServe())
				mutex.Unlock()
				slog.Error("server.ListenAndServe", "err", err)
			}
		}()
	}
	wg.Wait()
	// Wait a little to give the server time to start
	time.Sleep(1 * time.Second)
	return errs
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	var err error
	for i := range s.servers {
		err = errors.Join(err, s.servers[i].Shutdown(ctx))
	}
	return err
}
