package config

type Transport struct {
	DisableKeepAlives      bool     `yaml:"disableKeepAlives" json:"disableKeepAlives"`
	DisableCompression     bool     `yaml:"disableCompression" json:"disableCompression"`
	MaxIdleConns           int      `yaml:"maxIdleConns" json:"maxIdleConns"`
	MaxIdleConnsPerHost    int      `yaml:"maxIdleConnsPerHost" json:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int      `yaml:"maxConnsPerHost" json:"maxConnsPerHost"`
	IdleConnTimeout        Duration `yaml:"idleConnTimeout" json:"idleConnTimeout"`
	ResponseHeaderTimeout  Duration `yaml:"responseHeaderTimeout" json:"responseHeaderTimeout"`
	ExpectContinueTimeout  Duration `yaml:"expectContinueTimeout" json:"expectContinueTimeout"`
	MaxResponseHeaderBytes int64    `yaml:"maxResponseHeaderBytes" json:"maxResponseHeaderBytes"`
	WriteBufferSize        int      `yaml:"writeBufferSize" json:"writeBufferSize"`
	ReadBufferSize         int      `yaml:"readBufferSize" json:"readBufferSize"`
}
