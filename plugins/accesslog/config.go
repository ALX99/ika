package accesslog

type pConfig struct {
	// Headers contains the list of headers to log.
	Headers []string `json:"headers"`

	// IncludeRemoteAddr controls whether the remote address is included in the log.
	IncludeRemoteAddr bool `json:"logRemoteAddr"`
}
