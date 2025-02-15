package accesslog

type pConfig struct {
	// Headers is a list of headers to include in the log.
	Headers []string `json:"headers"`

	// RemoteAddr controls whether the remote address is included in the log.
	RemoteAddr bool `json:"remoteAddr"`
}

func (c *pConfig) SetDefaults() {
	if len(c.Headers) == 0 {
		c.Headers = []string{"X-Request-ID"}
	}
}
