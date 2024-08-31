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
	server http.Server
}

// NewServer creates a new server with the given handler and configuration.
func NewServer(handler http.Handler, config config.Server) *Server {
	return &Server{
		server: http.Server{
			Handler:                      handler,
			Addr:                         config.Addr,
			DisableGeneralOptionsHandler: config.DisableGeneralOptionsHandler,
			ReadTimeout:                  config.ReadTimeout,
			ReadHeaderTimeout:            config.ReadHeaderTimeout,
			WriteTimeout:                 config.WriteTimeout,
			IdleTimeout:                  config.IdleTimeout,
			MaxHeaderBytes:               config.MaxHeaderBytes,
		},
	}
}

// ListenAndServe starts the server and listens for incoming connections.
func (s *Server) ListenAndServe() error {
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		wg.Done()
		err = s.server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server.ListenAndServe", "err", err)
		}
	}()
	wg.Wait()
	// Wait a little to give the server time to start
	time.Sleep(1 * time.Second)
	return err
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return context.Cause(ctx)
		}
	}
	return err
}
