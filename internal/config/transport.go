package config

import "time"

type Transport struct {
	DisableKeepAlives      Nullable[bool]          `yaml:"disableKeepAlives"`
	DisableCompression     Nullable[bool]          `yaml:"disableCompression"`
	MaxIdleConns           Nullable[int]           `yaml:"maxIdleConns"`
	MaxIdleConnsPerHost    Nullable[int]           `yaml:"maxIdleConnsPerHost"`
	MaxConnsPerHost        Nullable[int]           `yaml:"maxConnsPerHost"`
	IdleConnTimeout        Nullable[time.Duration] `yaml:"idleConnTimeout"`
	ResponseHeaderTimeout  Nullable[time.Duration] `yaml:"responseHeaderTimeout"`
	ExpectContinueTimeout  Nullable[time.Duration] `yaml:"expectContinueTimeout"`
	MaxResponseHeaderBytes Nullable[int64]         `yaml:"maxResponseHeaderBytes"`
	WriteBufferSize        Nullable[int]           `yaml:"writeBufferSize"`
	ReadBufferSize         Nullable[int]           `yaml:"readBufferSize"`
}
