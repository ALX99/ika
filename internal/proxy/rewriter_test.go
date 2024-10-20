package proxy

import (
	"net/http"
	"net/url"
	"testing"
)

func Test_indexRewriter_rewrite(t *testing.T) {
	type args struct {
		r            *http.Request
		routePattern string
		isNamespaced bool
		toPattern    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "simple rewrite",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/bar"}},
				routePattern: "/{0}/bar",
				isNamespaced: false,
				toPattern:    "/baz/{0}",
			},
			want: "/baz/foo",
		},
		{
			name: "rewrite with namespace",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/ns/foo/bar"}},
				routePattern: "/{0}/bar",
				isNamespaced: true,
				toPattern:    "/baz/{0}",
			},
			want: "/baz/foo",
		},
		{
			name: "rewrite with wildcard",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/bar/baz"}},
				routePattern: "/{0}/bar/{1}",
				isNamespaced: false,
				toPattern:    "/baz/{0}/{1}",
			},
			want: "/baz/foo/baz",
		},
		{
			name: "rewrite with encoded path",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo%20bar/baz"}},
				routePattern: "/{0}/baz",
				isNamespaced: false,
				toPattern:    "/baz/{0}",
			},
			want: "/baz/foo%20bar",
		},
		{
			name: "rewrite with multiple segments",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/bar/baz/qux"}},
				routePattern: "/{0}/bar/{1}/qux",
				isNamespaced: false,
				toPattern:    "/new/{0}/{1}",
			},
			want: "/new/foo/baz",
		},
		{
			name: "rewrite with wildcard segment",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/bar/baz/qux"}},
				routePattern: "/{0}/bar/{1...}",
				isNamespaced: false,
				toPattern:    "/new/{0}/{1...}",
			},
			want: "/new/foo/baz/qux",
		},
		{
			name: "rewrite with wildcard segment and namespace",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/ns/foo/bar/baz/qux"}},
				routePattern: "/{0}/bar/{1...}",
				isNamespaced: true,
				toPattern:    "/new/{0}/{1...}",
			},
			want: "/new/foo/baz/qux",
		},
		{
			name: "trailing slash are preserved",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/bar/"}},
				routePattern: "/{0}/bar/",
				isNamespaced: false,
				toPattern:    "/baz/{0}/",
			},
			want: "/baz/foo/",
		},
		{
			name: "trailing slash is added",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo"}},
				routePattern: "/{0}",
				isNamespaced: false,
				toPattern:    "/baz/{0}/",
			},
			want: "/baz/foo/",
		},
		{
			name: "trailingslash are removed 1",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo/"}},
				routePattern: "/{0}",
				isNamespaced: false,
				toPattern:    "/baz/{0}",
			},
			want: "/baz/foo",
		},
		{
			name: "trailingslash are removed 2",
			args: args{
				r:            &http.Request{URL: &url.URL{Path: "/foo"}},
				routePattern: "/{0}",
				isNamespaced: false,
				toPattern:    "/baz/{0}",
			},
			want: "/baz/foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := newIndexRewriter(tt.args.routePattern, tt.args.isNamespaced, tt.args.toPattern)
			if got := ar.rewrite(tt.args.r); got != tt.want {
				t.Errorf("indexRewriter.rewrite() = %v, want %v", got, tt.want)
			}
		})
	}
}
