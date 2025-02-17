package basicauth

import (
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

//nolint:tparallel // not possible with t.Setenv
func TestPlugin_Setup(t *testing.T) {
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")

	factory := Factory()

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
		check   func(*is.I, ika.Plugin)
	}{
		{
			name: "valid static config with incoming",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "user",
							"password": "pass",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid static config with outgoing",
			config: map[string]any{
				"outgoing": map[string]any{
					"type":     "static",
					"username": "user",
					"password": "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "valid env config",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "env",
							"username": "USER_ENV",
							"password": "PASS_ENV",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "invalid",
							"username": "user",
							"password": "pass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"password": "pass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "user",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "env username not set",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "env",
							"username": "MISSING_USER_ENV",
							"password": "PASS_ENV",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "env password not set",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "env",
							"username": "USER_ENV",
							"password": "MISSING_PASS_ENV",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name:    "no config",
			config:  map[string]any{},
			wantErr: true,
		},
		{
			name: "both incoming and outgoing",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "user",
							"password": "pass",
						},
					},
				},
				"outgoing": map[string]any{
					"type":     "static",
					"username": "user",
					"password": "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "default type is set to static",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"username": "user",
							"password": "pass",
						},
					},
				},
				"outgoing": map[string]any{
					"username": "user",
					"password": "pass",
				},
			},
			wantErr: false,
			check: func(is *is.I, p ika.Plugin) {
				// We need to type assert here since we're checking internal state
				plugin := p.(*plugin)
				is.Equal(len(plugin.inCreds), 1)
				is.Equal(string(plugin.inCreds[0].user), "user")
				is.Equal(string(plugin.inCreds[0].pass), "pass")
				is.Equal(plugin.outUser, "user")
				is.Equal(plugin.outPass, "pass")
			},
		},
		{
			name: "multiple credentials",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "admin",
							"password": "adminpass",
						},
						{
							"name":     "user",
							"type":     "static",
							"username": "user",
							"password": "userpass",
						},
					},
				},
			},
			wantErr: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(len(plugin.inCreds), 2)
				is.Equal(string(plugin.inCreds[0].user), "admin")
				is.Equal(string(plugin.inCreds[0].pass), "adminpass")
				is.Equal(string(plugin.inCreds[1].user), "user")
				is.Equal(string(plugin.inCreds[1].pass), "userpass")
			},
		},
		{
			name: "duplicate credential names",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "admin",
							"password": "adminpass",
						},
						{
							"name":     "admin",
							"type":     "static",
							"username": "user",
							"password": "userpass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid static config with incoming and strip",
			config: map[string]any{
				"incoming": map[string]any{
					"credentials": []map[string]any{
						{
							"name":     "admin",
							"type":     "static",
							"username": "user",
							"password": "pass",
						},
					},
					"strip": true,
				},
			},
			wantErr: false,
			check: func(is *is.I, p ika.Plugin) {
				plugin := p.(*plugin)
				is.Equal(len(plugin.inCreds), 1)
				is.Equal(string(plugin.inCreds[0].user), "user")
				is.Equal(string(plugin.inCreds[0].pass), "pass")
				is.True(plugin.strip)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			p, err := factory.New(t.Context(), ika.InjectionContext{}, tt.config)

			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
				if tt.check != nil {
					tt.check(is, p)
				}
			}
		})
	}
}

func TestIncomingConfig_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     incomingConfig
		wantErr bool
	}{
		{
			name: "valid config with single credential",
			cfg: incomingConfig{
				Credentials: []namedCredential{
					{
						Name: "admin",
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid config with multiple credentials",
			cfg: incomingConfig{
				Credentials: []namedCredential{
					{
						Name: "admin",
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "admin",
							Password: "adminpass",
						},
					},
					{
						Name: "user",
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "user",
							Password: "userpass",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "no credentials",
			cfg: incomingConfig{
				Credentials: []namedCredential{},
			},
			wantErr: true,
		},
		{
			name: "missing credential name",
			cfg: incomingConfig{
				Credentials: []namedCredential{
					{
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "user",
							Password: "pass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "duplicate credential names",
			cfg: incomingConfig{
				Credentials: []namedCredential{
					{
						Name: "admin",
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "admin",
							Password: "adminpass",
						},
					},
					{
						Name: "admin",
						basicAuthConfig: basicAuthConfig{
							Type:     "static",
							Username: "user",
							Password: "userpass",
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid credential config",
			cfg: incomingConfig{
				Credentials: []namedCredential{
					{
						Name: "admin",
						basicAuthConfig: basicAuthConfig{
							Type: "invalid",
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			err := tt.cfg.validate()
			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
			}
		})
	}
}

//nolint:tparallel // not possible with t.Setenv
func TestNamedCredential_Validate(t *testing.T) {
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")

	tests := []struct {
		name    string
		cfg     namedCredential
		wantErr bool
	}{
		{
			name: "valid static config",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "static",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
		},
		{
			name: "valid env config",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "USER_ENV",
					Password: "PASS_ENV",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "invalid",
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "static",
					Password: "pass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "static",
					Username: "user",
				},
			},
			wantErr: true,
		},
		{
			name: "env username not set",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "MISSING_USER_ENV",
					Password: "PASS_ENV",
				},
			},
			wantErr: true,
		},
		{
			name: "env password not set",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "USER_ENV",
					Password: "MISSING_PASS_ENV",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			err := tt.cfg.validate()
			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
			}
		})
	}
}

//nolint:tparallel // not possible with t.Setenv
func TestBasicAuthConfig_Validate(t *testing.T) {
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")

	tests := []struct {
		name    string
		cfg     basicAuthConfig
		wantErr bool
	}{
		{
			name: "valid static config",
			cfg: basicAuthConfig{
				Type:     "static",
				Username: "user",
				Password: "pass",
			},
			wantErr: false,
		},
		{
			name: "valid env config",
			cfg: basicAuthConfig{
				Type:     "env",
				Username: "USER_ENV",
				Password: "PASS_ENV",
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			cfg: basicAuthConfig{
				Type:     "invalid",
				Username: "user",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing username",
			cfg: basicAuthConfig{
				Type:     "static",
				Password: "pass",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			cfg: basicAuthConfig{
				Type:     "static",
				Username: "user",
			},
			wantErr: true,
		},
		{
			name: "env username not set",
			cfg: basicAuthConfig{
				Type:     "env",
				Username: "MISSING_USER_ENV",
				Password: "PASS_ENV",
			},
			wantErr: true,
		},
		{
			name: "env password not set",
			cfg: basicAuthConfig{
				Type:     "env",
				Username: "USER_ENV",
				Password: "MISSING_PASS_ENV",
			},
			wantErr: true,
		},
		{
			name: "empty type defaults to static",
			cfg: basicAuthConfig{
				Username: "user",
				Password: "pass",
			},
			wantErr: true, // still fails because type is empty at validation time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			err := tt.cfg.validate()
			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
			}
		})
	}
}

//nolint:tparallel // not possible with t.Setenv
func TestCredentials(t *testing.T) {
	t.Setenv("USER_ENV", "envuser")
	t.Setenv("PASS_ENV", "envpass")

	tests := []struct {
		name       string
		cfg        namedCredential
		wantUser   string
		wantPass   string
		wantErr    bool
		setupEnv   func()
		cleanupEnv func()
	}{
		{
			name: "static credentials",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "static",
					Username: "staticuser",
					Password: "staticpass",
				},
			},
			wantUser: "staticuser",
			wantPass: "staticpass",
			wantErr:  false,
		},
		{
			name: "env credentials",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "USER_ENV",
					Password: "PASS_ENV",
				},
			},
			wantUser: "envuser",
			wantPass: "envpass",
			wantErr:  false,
		},
		{
			name: "missing env username",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "MISSING_USER_ENV",
					Password: "PASS_ENV",
				},
			},
			wantErr: true,
		},
		{
			name: "missing env password",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "USER_ENV",
					Password: "MISSING_PASS_ENV",
				},
			},
			wantErr: true,
		},
		{
			name: "env variables change during runtime",
			cfg: namedCredential{
				Name: "admin",
				basicAuthConfig: basicAuthConfig{
					Type:     "env",
					Username: "DYNAMIC_USER_ENV",
					Password: "DYNAMIC_PASS_ENV",
				},
			},
			setupEnv: func() {
				t.Setenv("DYNAMIC_USER_ENV", "dynamicuser")
				t.Setenv("DYNAMIC_PASS_ENV", "dynamicpass")
			},
			cleanupEnv: func() {
				t.Setenv("DYNAMIC_USER_ENV", "")
				t.Setenv("DYNAMIC_PASS_ENV", "")
			},
			wantUser: "dynamicuser",
			wantPass: "dynamicpass",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			if tt.setupEnv != nil {
				tt.setupEnv()
			}
			if tt.cleanupEnv != nil {
				defer tt.cleanupEnv()
			}

			user, pass, err := tt.cfg.credentials()
			if tt.wantErr {
				is.True(err != nil)
			} else {
				is.NoErr(err)
				is.Equal(user, tt.wantUser)
				is.Equal(pass, tt.wantPass)
			}
		})
	}
}
