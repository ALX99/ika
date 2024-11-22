package config

import "time"

type Server struct {
	Addr                         string                  `yaml:"addr"`
	DisableGeneralOptionsHandler Nullable[bool]          `yaml:"disableGeneralOptionsHandler"`
	ReadTimeout                  Nullable[time.Duration] `yaml:"readTimeout"`
	ReadHeaderTimeout            Nullable[time.Duration] `yaml:"readHeaderTimeout"`
	WriteTimeout                 Nullable[time.Duration] `yaml:"writeTimeout"`
	IdleTimeout                  Nullable[time.Duration] `yaml:"idleTimeout"`
	MaxHeaderBytes               Nullable[int]           `yaml:"maxHeaderBytes"`
}
