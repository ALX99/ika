package accesslog

type pConfig struct {
	// Headers is a list of headers to include in the log.
	Headers []string `json:"headers"`

	// RemoteAddr controls whether the remote address is included in the log.
	RemoteAddr bool `json:"remoteAddr"`

	// QueryParams is a list of query parameters to include in the log.
	// If empty, no query parameters will be logged.
	QueryParams []string `json:"queryParams"`
}

func (c *pConfig) SetDefaults() {
	if len(c.Headers) == 0 {
		c.Headers = []string{"X-Request-ID"}
	}
}
