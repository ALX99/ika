package requestid

import (
	"net/url"
	"testing"

	"github.com/matryer/is"
)

func Test_basicAuthConfig_validate(t *testing.T) {
	is := is.New(t)
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")
	type fields struct {
		Type     string
		Encoding string
		Username string
		Password string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid static config",
			fields: fields{
				Type:     "static",
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "valid env config",
			fields: fields{
				Type:     "env",
				Username: "USER_ENV",
				Password: "PASS_ENV",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			fields: fields{
				Type:     "invalid",
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			fields: fields{
				Type:     "static",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			fields: fields{
				Type:     "static",
				Username: "user",
			},
			wantErr: true,
		},
		{
			name: "env username not set",
			fields: fields{
				Type:     "env",
				Username: "MISSING_USER_ENV",
				Password: "PASS_ENV",
			},
			wantErr: true,
		},
		{
			name: "env password not set",
			fields: fields{
				Type:     "env",
				Username: "USER_ENV",
				Password: "MISSING_PASS_ENV",
			},
			wantErr: true,
		},
		{
			name: "invalid encoding",
			fields: fields{
				Type:     "static",
				Username: "user",
				Password: "pass",
				Encoding: "invalid",
			},
			wantErr: true,
		},
		{
			name: "username contains colon without encoding",
			fields: fields{
				Type:     "static",
				Username: "user:colon",
				Password: "pass",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &basicAuthConfig{
				Type:     tt.fields.Type,
				Encoding: tt.fields.Encoding,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			if err := c.validate(); !tt.wantErr {
				is.NoErr(err)
			} else {
				is.True(err != nil)
			}
		})
	}
}

func Test_basicAuthConfig_credentials(t *testing.T) {
	is := is.New(t)
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")
	t.Setenv("ENCODED_USER_ENV", url.QueryEscape("user:colon"))
	t.Setenv("ENCODED_PASS_ENV", url.QueryEscape("pass:colon"))

	type fields struct {
		Type     string
		Encoding string
		Username string
		Password string
	}
	tests := []struct {
		name     string
		fields   fields
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name: "static credentials",
			fields: fields{
				Type:     "static",
				Username: "user",
				Password: "pass",
			},
			wantUser: "user",
			wantPass: "pass",
			wantErr:  false,
		},
		{
			name: "env credentials",
			fields: fields{
				Type:     "env",
				Username: "USER_ENV",
				Password: "PASS_ENV",
			},
			wantUser: "user",
			wantPass: "pass",
			wantErr:  false,
		},
		{
			name: "env credentials with urlencoding",
			fields: fields{
				Type:     "env",
				Username: "ENCODED_USER_ENV",
				Password: "ENCODED_PASS_ENV",
				Encoding: "urlencoding",
			},
			wantUser: "user:colon",
			wantPass: "pass:colon",
			wantErr:  false,
		},
		{
			name: "missing env username",
			fields: fields{
				Type:     "env",
				Username: "MISSING_USER_ENV",
				Password: "PASS_ENV",
			},
			wantUser: "",
			wantPass: "",
			wantErr:  true,
		},
		{
			name: "missing env password",
			fields: fields{
				Type:     "env",
				Username: "USER_ENV",
				Password: "MISSING_PASS_ENV",
			},
			wantUser: "",
			wantPass: "",
			wantErr:  true,
		},
		{
			name: "failed to unescape username",
			fields: fields{
				Type:     "env",
				Username: "INVALID_ENCODED_USER_ENV",
				Password: "PASS_ENV",
				Encoding: "urlencoding",
			},
			wantUser: "",
			wantPass: "",
			wantErr:  true,
		},
		{
			name: "failed to unescape password",
			fields: fields{
				Type:     "env",
				Username: "USER_ENV",
				Password: "INVALID_ENCODED_PASS_ENV",
				Encoding: "urlencoding",
			},
			wantUser: "",
			wantPass: "",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &basicAuthConfig{
				Type:     tt.fields.Type,
				Encoding: tt.fields.Encoding,
				Username: tt.fields.Username,
				Password: tt.fields.Password,
			}
			gotUser, gotPass, err := c.credentials()
			if !tt.wantErr {
				is.NoErr(err)
			} else {
				is.True(err != nil)
			}
			is.Equal(gotUser, tt.wantUser)
			is.Equal(gotPass, tt.wantPass)
		})
	}
}

func Test_pConfig_validate(t *testing.T) {
	is := is.New(t)
	tests := []struct {
		name    string
		config  pConfig
		wantErr bool
	}{
		{
			name: "valid config with incoming",
			config: pConfig{
				Incoming: &basicAuthConfig{
					Type:     "static",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with outgoing",
			config: pConfig{
				Outgoing: &basicAuthConfig{
					Type:     "static",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with both incoming and outgoing",
			config: pConfig{
				Incoming: &basicAuthConfig{
					Type:     "static",
					Username: "user",
					Password: "pass",
				},
				Outgoing: &basicAuthConfig{
					Type:     "static",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid config with neither incoming nor outgoing",
			config:  pConfig{},
			wantErr: true,
		},
		{
			name: "invalid incoming config",
			config: pConfig{
				Incoming: &basicAuthConfig{
					Type:     "invalid",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid outgoing config",
			config: pConfig{
				Outgoing: &basicAuthConfig{
					Type:     "invalid",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if !tt.wantErr {
				is.NoErr(err)
			} else {
				is.True(err != nil)
			}
		})
	}
}
