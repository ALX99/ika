package router

import (
	"testing"

	"github.com/alx99/ika/internal/config"
	"github.com/stretchr/testify/assert"
)

func Test_makeRoutePattern(t *testing.T) {
	type args struct {
		routePattern string
		nsName       string
		ns           config.Namespace
		route        config.Path
	}
	tests := []struct {
		name string
		args args
		want []routePattern
	}{
		{
			name: "Test with hosts",
			args: args{
				routePattern: "/",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{
				{pattern: "GET host1.com/"},
				{pattern: "GET host2.com/"},
				{pattern: "POST host1.com/"},
				{pattern: "POST host2.com/"},
				{pattern: "GET /nsName/", isNamespaced: true},
				{pattern: "POST /nsName/", isNamespaced: true},
			},
		},
		{
			"test with route pattern",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			[]routePattern{
				{pattern: "GET host1.com/route/{something/{any...}}"},
				{pattern: "GET host2.com/route/{something/{any...}}"},
				{pattern: "POST host1.com/route/{something/{any...}}"},
				{pattern: "POST host2.com/route/{something/{any...}}"},
				{pattern: "GET /nsName/route/{something/{any...}}", isNamespaced: true},
				{pattern: "POST /nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
		{
			"test with disable namespace routes",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					Name:                   "nsName",
					DisableNamespacedPaths: config.NewNullable(true),
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			[]routePattern{},
		},
		{
			"test with empty hosts",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			[]routePattern{
				{pattern: "GET /nsName/route/{something/{any...}}", isNamespaced: true},
				{pattern: "POST /nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeRoutes(tt.args.routePattern, tt.args.ns, tt.args.route)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
