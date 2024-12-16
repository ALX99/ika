package config

import "time"

type Server struct {
	Addr                         string        `yaml:"addr"`
	DisableGeneralOptionsHandler bool          `yaml:"disableGeneralOptionsHandler"`
	ReadTimeout                  time.Duration `yaml:"readTimeout"`
	ReadHeaderTimeout            time.Duration `yaml:"readHeaderTimeout"`
	WriteTimeout                 time.Duration `yaml:"writeTimeout"`
	IdleTimeout                  time.Duration `yaml:"idleTimeout"`
	MaxHeaderBytes               int           `yaml:"maxHeaderBytes"`
}
