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
			name: "Test with route pattern",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{
				{pattern: "GET host1.com/route/{something/{any...}}"},
				{pattern: "GET host2.com/route/{something/{any...}}"},
				{pattern: "POST host1.com/route/{something/{any...}}"},
				{pattern: "POST host2.com/route/{something/{any...}}"},
				{pattern: "GET /nsName/route/{something/{any...}}", isNamespaced: true},
				{pattern: "POST /nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
		{
			name: "Test with disable namespace routes",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:                   "nsName",
					DisableNamespacedPaths: config.NewNullable(true),
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{},
		},
		{
			name: "Test with empty hosts",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{
				{pattern: "GET /nsName/route/{something/{any...}}", isNamespaced: true},
				{pattern: "POST /nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
		{
			name: "Test with no methods",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{},
				},
			},
			want: []routePattern{
				{pattern: "host1.com/route/{something/{any...}}"},
				{pattern: "host2.com/route/{something/{any...}}"},
				{pattern: "/nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
		{
			name: "Test with no hosts and no methods",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{},
				},
				route: config.Path{
					Methods: []config.Method{},
				},
			},
			want: []routePattern{
				{pattern: "/nsName/route/{something/{any...}}", isNamespaced: true},
			},
		},
		{
			name: "Test with root namespace",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "root",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{
				{pattern: "GET host1.com/route/{something/{any...}}"},
				{pattern: "GET host2.com/route/{something/{any...}}"},
				{pattern: "POST host1.com/route/{something/{any...}}"},
				{pattern: "POST host2.com/route/{something/{any...}}"},
				{pattern: "GET /route/{something/{any...}}", isNamespaced: false},
				{pattern: "POST /route/{something/{any...}}", isNamespaced: false},
			},
		},
		{
			name: "Test with root namespace and no methods",
			args: args{
				routePattern: "/route/{something/{any...}}",
				ns: config.Namespace{
					Name:  "root",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{},
				},
			},
			want: []routePattern{
				{pattern: "host1.com/route/{something/{any...}}"},
				{pattern: "host2.com/route/{something/{any...}}"},
				{pattern: "/route/{something/{any...}}", isNamespaced: false},
			},
		},
		{
			name: "Test with empty route pattern",
			args: args{
				routePattern: "",
				ns: config.Namespace{
					Name:  "nsName",
					Hosts: []string{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			want: []routePattern{
				{pattern: "GET /nsName", isNamespaced: true},
				{pattern: "POST /nsName", isNamespaced: true},
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
