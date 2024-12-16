package config

type Server struct {
	Addr                         string   `yaml:"addr" json:"addr"`
	DisableGeneralOptionsHandler bool     `yaml:"disableGeneralOptionsHandler" json:"disableGeneralOptionsHandler"`
	ReadTimeout                  Duration `yaml:"readTimeout" json:"readTimeout"`
	ReadHeaderTimeout            Duration `yaml:"readHeaderTimeout" json:"readHeaderTimeout"`
	WriteTimeout                 Duration `yaml:"writeTimeout" json:"writeTimeout"`
	IdleTimeout                  Duration `yaml:"idleTimeout" json:"idleTimeout"`
	MaxHeaderBytes               int      `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`
}
