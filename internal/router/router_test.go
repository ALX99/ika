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
		want []string
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
			want: []string{
				"GET host1.com/",
				"GET host2.com/",
				"POST host1.com/",
				"POST host2.com/",
				"GET /nsName/",
				"POST /nsName/",
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
			[]string{
				"GET host1.com/route/{something/{any...}}",
				"GET host2.com/route/{something/{any...}}",
				"POST host1.com/route/{something/{any...}}",
				"POST host2.com/route/{something/{any...}}",
				"GET /nsName/route/{something/{any...}}",
				"POST /nsName/route/{something/{any...}}",
			},
		},
		{
			"test with disable namespace routes",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					Name:                   "nsName",
					DisableNamespacedPaths: true,
				},
				route: config.Path{
					Methods: []config.Method{"GET", "POST"},
				},
			},
			[]string{},
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
			[]string{
				"GET /nsName/route/{something/{any...}}",
				"POST /nsName/route/{something/{any...}}",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := makeRoutePatterns(tt.args.routePattern, tt.args.ns, tt.args.route)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
