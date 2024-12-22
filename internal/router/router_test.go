package router

import (
	"testing"

	"github.com/alx99/ika/internal/config"
	"github.com/gkampitakis/go-snaps/snaps"
)

func Test_makeRoutes(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		rp     string
		nsName string
		route  config.Path
	}{
		{
			name:   "root namespace with no methods",
			rp:     "/test",
			nsName: "root",
			route:  config.Path{},
		},
		{
			name:   "namespaced route with no methods",
			rp:     "/test",
			nsName: "/namespace",
			route:  config.Path{},
		},
		{
			name:   "host route with no methods",
			rp:     "/test",
			nsName: "example.com",
			route:  config.Path{},
		},
		{
			name:   "root namespace with methods",
			rp:     "/test",
			nsName: "root",
			route:  config.Path{Methods: []config.Method{"GET", "POST"}},
		},
		{
			name:   "namespaced route with methods",
			rp:     "/test",
			nsName: "/namespace",
			route:  config.Path{Methods: []config.Method{"GET", "POST"}},
		},
		{
			name:   "host route with methods",
			rp:     "/test",
			nsName: "example.com",
			route:  config.Path{Methods: []config.Method{"GET", "POST"}},
		},
		{
			name:   "host with empty route",
			rp:     "",
			nsName: "example.com",
			route:  config.Path{Methods: []config.Method{"GET", "POST"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeRoutes(tt.rp, tt.nsName, tt.route)
			snaps.MatchSnapshot(t, got)
		})
	}
}
