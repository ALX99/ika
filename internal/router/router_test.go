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
				nsName:       "nsName",
				ns: config.Namespace{
					Hosts: []config.Host{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []string{"GET", "POST"},
				},
			},
			want: []string{
				"GET host1.com/",
				"GET host2.com/",
				"POST host1.com/",
				"POST host2.com/",
			},
		},
		{
			"test with route pattern",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					Hosts: []config.Host{"host1.com", "host2.com"},
				},
				route: config.Path{
					Methods: []string{"GET", "POST"},
				},
			},
			[]string{
				"GET host1.com/route/{something/{any...}}",
				"GET host2.com/route/{something/{any...}}",
				"POST host1.com/route/{something/{any...}}",
				"POST host2.com/route/{something/{any...}}",
			},
		},
		{
			"test with disable namespace routes",
			args{
				routePattern: "/route/{something/{any...}}",
				nsName:       "nsName",
				ns: config.Namespace{
					DisableNaspacmedRoutes: true,
				},
				route: config.Path{
					Methods: []string{"GET", "POST"},
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
					Hosts: []config.Host{},
				},
				route: config.Path{
					Methods: []string{"GET", "POST"},
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
			got := makeRoutePattern(tt.args.routePattern, tt.args.nsName, tt.args.ns, tt.args.route)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}
