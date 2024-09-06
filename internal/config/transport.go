package config

import (
	"time"
)

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

// override returns a new Transport with the values of t2 overriding the values of t1.
func override(t1, t2 Transport) Transport {
	return Transport{
		DisableKeepAlives:      firstSet(t2.DisableKeepAlives, t1.DisableKeepAlives),
		DisableCompression:     firstSet(t2.DisableCompression, t1.DisableCompression),
		MaxIdleConns:           firstSet(t2.MaxIdleConns, t1.MaxIdleConns),
		MaxIdleConnsPerHost:    firstSet(t2.MaxIdleConnsPerHost, t1.MaxIdleConnsPerHost),
		MaxConnsPerHost:        firstSet(t2.MaxConnsPerHost, t1.MaxConnsPerHost),
		IdleConnTimeout:        firstSet(t2.IdleConnTimeout, t1.IdleConnTimeout),
		ResponseHeaderTimeout:  firstSet(t2.ResponseHeaderTimeout, t1.ResponseHeaderTimeout),
		ExpectContinueTimeout:  firstSet(t2.ExpectContinueTimeout, t1.ExpectContinueTimeout),
		MaxResponseHeaderBytes: firstSet(t2.MaxResponseHeaderBytes, t1.MaxResponseHeaderBytes),
		WriteBufferSize:        firstSet(t2.WriteBufferSize, t1.WriteBufferSize),
		ReadBufferSize:         firstSet(t2.ReadBufferSize, t1.ReadBufferSize),
	}
}
