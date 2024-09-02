package config

import "time"

type Transport struct {
	DisableKeepAlives      bool          `yaml:"disableKeepAlives"`
	DisableCompression     bool          `yaml:"disableCompression"`
	MaxIdleConns           int           `yaml:"maxIdleConns"`
	MaxIdleConnsPerHost    int           `yaml:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int           `yaml:"maxConnsPerHost"`
	IdleConnTimeout        time.Duration `yaml:"idleConnTimeout"`
	ResponseHeaderTimeout  time.Duration `yaml:"responseHeaderTimeout"`
	ExpectContinueTimeout  time.Duration `yaml:"expectContinueTimeout"`
	MaxResponseHeaderBytes int64         `yaml:"maxResponseHeaderBytes"`
	WriteBufferSize        int           `yaml:"writeBufferSize"`
	ReadBufferSize         int           `yaml:"readBufferSize"`
}
