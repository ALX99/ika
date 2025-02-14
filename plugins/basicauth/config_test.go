package basicauth

import (
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func TestPlugin_Setup(t *testing.T) {
	t.Setenv("USER_ENV", "user")
	t.Setenv("PASS_ENV", "pass")

	tests := []struct {
		name    string
		config  map[string]any
		wantErr bool
		check   func(*is.I, *Plugin)
	}{
		{
			name: "valid static config with incoming",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "static",
					"username": "user",
					"password": "pass",
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
					"type":     "env",
					"username": "USER_ENV",
					"password": "PASS_ENV",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid type",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "invalid",
					"username": "user",
					"password": "pass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing username",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "static",
					"password": "pass",
				},
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "static",
					"username": "user",
				},
			},
			wantErr: true,
		},
		{
			name: "env username not set",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "env",
					"username": "MISSING_USER_ENV",
					"password": "PASS_ENV",
				},
			},
			wantErr: true,
		},
		{
			name: "env password not set",
			config: map[string]any{
				"incoming": map[string]any{
					"type":     "env",
					"username": "USER_ENV",
					"password": "MISSING_PASS_ENV",
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
					"type":     "static",
					"username": "user",
					"password": "pass",
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
					"username": "user",
					"password": "pass",
				},
				"outgoing": map[string]any{
					"username": "user",
					"password": "pass",
				},
			},
			wantErr: false,
			check: func(is *is.I, p *Plugin) {
				// Check that credentials were set correctly, which means the default type "static" was used
				is.Equal(string(p.inUser), "user")
				is.Equal(string(p.inPass), "pass")
				is.Equal(p.outUser, "user")
				is.Equal(p.outPass, "pass")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)
			p := &Plugin{}
			err := p.Setup(t.Context(), ika.InjectionContext{}, tt.config)

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
