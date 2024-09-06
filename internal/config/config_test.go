package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestApplyOverride(t *testing.T) {
	tests := []struct {
		name     string
		orig     Transport
		override Transport
	}{
		{
			name: "Test override transport 1",
			orig: Transport{
				DisableKeepAlives: NewNullable(true),
				IdleConnTimeout:   NewNullable(time.Second),
			},
			override: Transport{
				DisableKeepAlives: NewNullable(false),
				IdleConnTimeout:   NewNullable(time.Minute),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Namespaces: Namespaces{
					"test": Namespace{
						Transport: tt.orig,
					},
				},
				NamespaceOverride: Namespace{
					Transport: tt.override,
				},
			}
			cfg.ApplyOverride()

			assert.Equal(t, tt.override, cfg.Namespaces["test"].Transport)
		})
	}
}
