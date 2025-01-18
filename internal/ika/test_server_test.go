package ika

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/http/server"
)

var _ server.HTTPServer = &testServer{}

type testServer struct {
	server  *httptest.Server
	startCh chan struct{}
}

func (s *testServer) ListenAndServe() error {
	s.server.Start()
	s.startCh <- struct{}{}
	return nil
}

func (s *testServer) Shutdown(context.Context) error {
	return nil
}

func (s *testServer) waitForStart() {
	<-s.startCh
}

func newTestServer(t *testing.T, s *httptest.Server) *testServer {
	t.Cleanup(s.Close)
	return &testServer{
		server:  s,
		startCh: make(chan struct{}, 1),
	}
}

func newMakeTestServer(t *testing.T, s *testServer) func(handler http.Handler, servers []config.Server) server.HTTPServer {
	return func(handler http.Handler, servers []config.Server) server.HTTPServer {
		s.server.Config.Handler = handler
		if len(servers) > 1 {
			t.Fatalf("test server does not support multiple servers")
		}
		if len(servers) == 1 {
			s.server.Config = server.ConfigureServer(s.server.Config, servers[0])
		}
		return s
	}
}
