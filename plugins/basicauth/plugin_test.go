package requestid

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func Test_plugin_ServeHTTP(t *testing.T) {
	is := is.New(t)
	type fields struct {
		inUser     []byte
		inPass     []byte
		outUser    string
		outPass    string
		inEncoding string
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantOutUser string
		wantOutPass string
	}{
		{
			name: "no incoming credentials",
			fields: fields{
				inUser:  nil,
				inPass:  nil,
				outUser: "outUser",
				outPass: "outPass",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
			},
			wantErr:     false,
			wantOutUser: "outUser",
			wantOutPass: "outPass",
		},
		{
			name: "invalid incoming credentials",
			fields: fields{
				inUser:  []byte("user"),
				inPass:  []byte("pass"),
				outUser: "",
				outPass: "",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
			},
			wantErr: true,
		},
		{
			name: "valid incoming credentials",
			fields: fields{
				inUser:  []byte("user"),
				inPass:  []byte("pass"),
				outUser: "outUser",
				outPass: "outPass",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req := httptest.NewRequest("GET", "/", nil)
					req.SetBasicAuth("user", "pass")
					return req
				}(),
			},
			wantErr:     false,
			wantOutUser: "outUser",
			wantOutPass: "outPass",
		},
		{
			name: "valid incoming credentials with urlencoding",
			fields: fields{
				inUser:     []byte("user:colon"),
				inPass:     []byte("pass:colon"),
				outUser:    "outUser",
				outPass:    "outPass",
				inEncoding: "urlencoding",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req := httptest.NewRequest("GET", "/", nil)
					req.SetBasicAuth(url.QueryEscape("user:colon"), url.QueryEscape("pass:colon"))
					return req
				}(),
			},
			wantErr:     false,
			wantOutUser: "outUser",
			wantOutPass: "outPass",
		},
		{
			name: "invalid incoming credentials",
			fields: fields{
				inUser:  []byte("user"),
				inPass:  []byte("pass"),
				outUser: "",
				outPass: "",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req := httptest.NewRequest("GET", "/", nil)
					req.SetBasicAuth("invalidUser", "invalidPass")
					return req
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &plugin{
				inUser:     tt.fields.inUser,
				inPass:     tt.fields.inPass,
				outUser:    tt.fields.outUser,
				outPass:    tt.fields.outPass,
				inEncoding: tt.fields.inEncoding,
				next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					return nil
				}),
			}
			err := p.ServeHTTP(tt.args.w, tt.args.r)
			if !tt.wantErr {
				is.NoErr(err)
			} else {
				is.True(err != nil)
			}
			user, pass, ok := tt.args.r.BasicAuth()
			if tt.wantOutUser != "" || tt.wantOutPass != "" {
				is.True(ok)
				is.Equal(user, tt.wantOutUser)
				is.Equal(pass, tt.wantOutPass)
			}
		})
	}
}
