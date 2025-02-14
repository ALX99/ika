package pluginutil

import (
	"errors"
	"testing"
	"time"

	"github.com/matryer/is"
)

type configWithDefaults struct {
	Duration time.Duration `json:"duration"`
}

func (c *configWithDefaults) SetDefaults() {
	if c.Duration == 0 {
		c.Duration = time.Hour
	}
}

type configWithValidation struct {
	Duration time.Duration `json:"duration"`
}

func (c *configWithValidation) Validate() error {
	if c.Duration < time.Minute {
		return errors.New("duration must be at least 1 minute")
	}
	return nil
}

func TestUnmarshalCfg_TimeParsing(t *testing.T) {
	t.Parallel()

	type config struct {
		Duration time.Duration `json:"duration"`
	}

	tests := []struct {
		name      string
		input     map[string]any
		want      config
		wantError bool
	}{
		{
			name: "valid duration and time",
			input: map[string]any{
				"duration": "1h30m",
			},
			want: config{
				Duration: 90 * time.Minute,
			},
		},
		{
			name: "valid with different units",
			input: map[string]any{
				"duration": "24h",
			},
			want: config{
				Duration: 24 * time.Hour,
			},
		},
		{
			name: "invalid duration",
			input: map[string]any{
				"duration": "invalid",
			},
			wantError: true,
		},
		{
			name: "int duration",
			input: map[string]any{
				"duration": 123,
			},
			want: config{
				Duration: time.Duration(123),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			var got config
			err := UnmarshalCfg(tt.input, &got)

			if tt.wantError {
				is.True(err != nil) // expected an error
				return
			}

			is.NoErr(err)                            // unexpected error
			is.Equal(got.Duration, tt.want.Duration) // duration matches
		})
	}
}

func TestUnmarshalCfg_TimeWithDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input map[string]any
		want  configWithDefaults
	}{
		{
			name:  "with defaults",
			input: map[string]any{},
			want: configWithDefaults{
				Duration: time.Hour,
			},
		},
		{
			name: "override defaults",
			input: map[string]any{
				"duration": "30m",
			},
			want: configWithDefaults{
				Duration: 30 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			var got configWithDefaults
			err := UnmarshalCfg(tt.input, &got)
			is.NoErr(err)                            // unexpected error
			is.Equal(got.Duration, tt.want.Duration) // duration matches
		})
	}
}

func TestUnmarshalCfg_TimeWithValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     map[string]any
		wantError bool
	}{
		{
			name: "valid config",
			input: map[string]any{
				"duration": "5m",
			},
		},
		{
			name: "invalid duration",
			input: map[string]any{
				"duration": "30s",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			is := is.New(t)

			var got configWithValidation
			err := UnmarshalCfg(tt.input, &got)

			if tt.wantError {
				is.True(err != nil) // expected an error
				return
			}

			is.NoErr(err) // unexpected error
		})
	}
}
