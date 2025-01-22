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
	Dialer                 Dialer   `json:"dialer"`
}

type Dialer struct {
	Timeout         Duration        `json:"timeout"`
	FallbackDelay   Duration        `json:"fallbackDelay"`
	KeepAliveConfig KeepAliveConfig `json:"keepAliveConfig"`
}

type KeepAliveConfig struct {
	Enable   bool     `json:"enable"`
	Idle     Duration `json:"idle"`
	Interval Duration `json:"interval"`
	Count    int      `json:"count"`
}
