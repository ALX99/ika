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

type HTTPServer interface {
	ListenAndServe() error
	Shutdown(context.Context) error
}

type MultiServer struct {
	servers []http.Server
}

func New(handler http.Handler, config []config.Server) *MultiServer {
	var servers []http.Server
	for _, c := range config {
		servers = append(servers, *ConfigureServer(&http.Server{Handler: handler}, c))
	}

	return &MultiServer{servers: servers}
}

func ConfigureServer(s *http.Server, c config.Server) *http.Server {
	s.Addr = c.Addr
	s.DisableGeneralOptionsHandler = c.DisableGeneralOptionsHandler
	s.ReadTimeout = c.ReadTimeout.Dur()
	s.ReadHeaderTimeout = c.ReadHeaderTimeout.Dur()
	s.WriteTimeout = c.WriteTimeout.Dur()
	s.IdleTimeout = c.IdleTimeout.Dur()
	s.MaxHeaderBytes = c.MaxHeaderBytes
	return s
}

func (s *MultiServer) ListenAndServe() error {
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

func (s *MultiServer) Shutdown(ctx context.Context) error {
	var err error
	for i := range s.servers {
		err = errors.Join(err, s.servers[i].Shutdown(ctx))
	}
	return err
}
