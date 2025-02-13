package accesslog

type pConfig struct {
	// Headers is a list of headers to include in the log.
	Headers []string `json:"headers"`

	// RemoteAddr controls whether the remote address is included in the log.
	RemoteAddr bool `json:"remoteAddr"`
}
