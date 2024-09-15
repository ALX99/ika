package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"testing"
)

func Test_setPath(t *testing.T) {
	type args struct {
		rp       *httputil.ProxyRequest
		rawPath  string
		wantPath string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Valid path rewrite",
			args: args{
				rp: &httputil.ProxyRequest{
					In: &http.Request{
						URL: &url.URL{},
					},
					Out: &http.Request{
						URL: &url.URL{
							RawPath: "/newpath",
							Path:    "/newpath",
						},
					},
				},
				rawPath:  "/newpath",
				wantPath: "/newpath",
			},
		},
		{
			name: "Unescaped path",
			args: args{
				rp: &httputil.ProxyRequest{
					In: &http.Request{
						URL: &url.URL{},
					},
					Out: &http.Request{
						URL: &url.URL{
							RawPath: "/old%20path",
							Path:    "/old path",
						},
					},
				},
				rawPath:  "/new%20path",
				wantPath: "/new path",
			},
		},
		{
			name: "Empty rawPath",
			args: args{
				rp: &httputil.ProxyRequest{
					In: &http.Request{
						URL: &url.URL{},
					},
					Out: &http.Request{
						URL: &url.URL{
							RawPath: "/oldpath",
							Path:    "/oldpath",
						},
					},
				},
				rawPath:  "",
				wantPath: "",
			},
		},
		{
			name: "Path with special characters",
			args: args{
				rp: &httputil.ProxyRequest{
					In: &http.Request{
						URL: &url.URL{},
					},
					Out: &http.Request{
						URL: &url.URL{
							RawPath: "/old%40path",
							Path:    "/old@path",
						},
					},
				},
				rawPath:  "/new%40path",
				wantPath: "/new@path",
			},
		},
		{
			name: "Path with query parameters",
			args: args{
				rp: &httputil.ProxyRequest{
					In: &http.Request{
						URL: &url.URL{},
					},
					Out: &http.Request{
						URL: &url.URL{
							RawPath: "/oldpath?query=1",
							Path:    "/oldpath",
						},
					},
				},
				rawPath:  "/newpath?query=1",
				wantPath: "/newpath",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setPath(tt.args.rp, tt.args.rawPath)
			if tt.args.rawPath != "" && tt.args.rawPath != tt.args.rp.Out.URL.RawPath {
				t.Errorf("Expected RawPath to be %s, got %s", tt.args.rawPath, tt.args.rp.Out.URL.RawPath)
			}
			if tt.args.wantPath != tt.args.rp.Out.URL.Path {
				t.Errorf("Expected Path to be %s, got %s", tt.args.wantPath, tt.args.rp.Out.URL.Path)
			}
		})
	}
}
