package basicauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func Test_plugin_ServeHTTP(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	type fields struct {
		inCreds []credential
		strip   bool
		outUser string
		outPass string
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
				inCreds: nil,
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
				inCreds: []credential{
					{
						user: []byte("user"),
						pass: []byte("pass"),
					},
				},
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
				inCreds: []credential{
					{
						user: []byte("user"),
						pass: []byte("pass"),
					},
				},
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
			name: "invalid incoming credentials",
			fields: fields{
				inCreds: []credential{
					{
						user: []byte("user"),
						pass: []byte("pass"),
					},
				},
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
		{
			name: "multiple valid credentials - first matches",
			fields: fields{
				inCreds: []credential{
					{
						user: []byte("admin"),
						pass: []byte("adminpass"),
					},
					{
						user: []byte("user"),
						pass: []byte("userpass"),
					},
				},
				outUser: "outUser",
				outPass: "outPass",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req := httptest.NewRequest("GET", "/", nil)
					req.SetBasicAuth("admin", "adminpass")
					return req
				}(),
			},
			wantErr:     false,
			wantOutUser: "outUser",
			wantOutPass: "outPass",
		},
		{
			name: "multiple valid credentials - second matches",
			fields: fields{
				inCreds: []credential{
					{
						user: []byte("admin"),
						pass: []byte("adminpass"),
					},
					{
						user: []byte("user"),
						pass: []byte("userpass"),
					},
				},
				outUser: "outUser",
				outPass: "outPass",
			},
			args: args{
				w: httptest.NewRecorder(),
				r: func() *http.Request {
					req := httptest.NewRequest("GET", "/", nil)
					req.SetBasicAuth("user", "userpass")
					return req
				}(),
			},
			wantErr:     false,
			wantOutUser: "outUser",
			wantOutPass: "outPass",
		},
		{
			name: "strip incoming credentials",
			fields: fields{
				inCreds: []credential{
					{
						user: []byte("user"),
						pass: []byte("pass"),
					},
				},
				strip:   true,
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
			name: "strip incoming credentials - no outgoing credentials",
			fields: fields{
				inCreds: []credential{
					{
						user: []byte("user"),
						pass: []byte("pass"),
					},
				},
				strip:   true,
				outUser: "",
				outPass: "",
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
			wantOutUser: "",
			wantOutPass: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &plugin{
				inCreds: tt.fields.inCreds,
				strip:   tt.fields.strip,
				outUser: tt.fields.outUser,
				outPass: tt.fields.outPass,
				next: ika.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
					return nil
				}),
			}
			err := p.ServeHTTP(tt.args.w, tt.args.r)
			if !tt.wantErr {
				is.NoErr(err)
			} else {
				is.True(err != nil)
				return // wanted error
			}

			user, pass, ok := tt.args.r.BasicAuth()
			is.True(ok == (tt.wantOutUser != "" || tt.wantOutPass != ""))
			if tt.wantOutUser != "" || tt.wantOutPass != "" {
				is.Equal(user, tt.wantOutUser)
				is.Equal(pass, tt.wantOutPass)
			}
		})
	}
}
