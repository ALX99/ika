package router

import (
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func Test_mergePaths(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		nsPath string
		path   string
	}{
		{
			name:   "root path join",
			nsPath: "/",
			path:   "/",
		},
		{
			name:   "normal path join",
			nsPath: "/test",
			path:   "/path",
		},
		{
			name:   "common base path",
			nsPath: "/test",
			path:   "/test/hi",
		},
		{
			name:   "with variables",
			nsPath: "/test/{something}",
			path:   "/test/hi/ok",
		},
		{
			name:   "with methods",
			nsPath: "POST /",
			path:   "/",
		},
		{
			name:   "with host",
			nsPath: "example.com:99/",
			path:   "/",
		},
		{
			name:   "with host and method",
			nsPath: "example.com/",
			path:   "GET /",
		},
		{
			name:   "with method and host",
			nsPath: "GET /example",
			path:   "example.com/",
		},
		{
			name:   "with mismatch methods",
			nsPath: "GET /example",
			path:   "POST /",
		},
		{
			name:   "with mismatch hosts",
			nsPath: "GET example.com/example",
			path:   "example1.com/",
		},
		{
			name:   "with matching hosts",
			nsPath: "GET example.com/example",
			path:   "example.com/",
		},
		{
			name:   "multiple slashes",
			nsPath: "GET //",
			path:   "//",
		},
		{
			name:   "tabbed path",
			nsPath: "TRACE /",
			path:   "TRACE\t/",
		},
		{
			name:   "spaces and tabs before path",
			nsPath: "GET /",
			path:   " \t /",
		},
		{
			name:   "path more specific",
			nsPath: "/",
			path:   "GET example.com/hi",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mergePaths(tt.nsPath, tt.path)
			snaps.MatchSnapshot(t, got, err)
		})
	}
}
