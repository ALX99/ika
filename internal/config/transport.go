package config

type Transport struct {
	DisableKeepAlives      bool     `json:"disableKeepAlives"`
	DisableCompression     bool     `json:"disableCompression"`
	MaxIdleConns           int      `json:"maxIdleConns"`
	MaxIdleConnsPerHost    int      `json:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int      `json:"maxConnsPerHost"`
	IdleConnTimeout        Duration `json:"idleConnTimeout"`
	ResponseHeaderTimeout  Duration `json:"responseHeaderTimeout"`
	ExpectContinueTimeout  Duration `json:"expectContinueTimeout"`
	MaxResponseHeaderBytes int64    `json:"maxResponseHeaderBytes"`
	WriteBufferSize        int      `json:"writeBufferSize"`
	ReadBufferSize         int      `json:"readBufferSize"`
}
