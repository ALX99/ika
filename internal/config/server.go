package config

type Server struct {
	Addr                         string   `json:"addr"`
	DisableGeneralOptionsHandler bool     `json:"disableGeneralOptionsHandler"`
	ReadTimeout                  Duration `json:"readTimeout"`
	ReadHeaderTimeout            Duration `json:"readHeaderTimeout"`
	WriteTimeout                 Duration `json:"writeTimeout"`
	IdleTimeout                  Duration `json:"idleTimeout"`
	MaxHeaderBytes               int      `json:"maxHeaderBytes"`
}
