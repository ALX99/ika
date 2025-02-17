package reqmodifier

import (
	"fmt"
	"net/url"
)

type pConfig struct {
	// Path is the path to rewrite to. It can contain segments from the original path
	// using {segment} syntax. For example: /new/{id}/path/{wildcard...}
	Path string `json:"path"`

	// Host is the host to rewrite to. Must be a valid URL including scheme.
	// For example: https://example.com
	Host string `json:"host"`

	// RetainHostHeader controls whether to keep the original Host header
	// when rewriting the host. If false, the Host header will be set to
	// the new host.
	RetainHostHeader bool `json:"retainHostHeader"`
}

func (c *pConfig) Validate() error {
	if c.Path == "" && c.Host == "" {
		return fmt.Errorf("at least one of path or host must be set")
	}

	if c.Host != "" {
		host := segmentPattern.ReplaceAllString(c.Host, "dummy")
		if _, err := url.Parse(host); err != nil {
			return fmt.Errorf("invalid host URL: %w", err)
		}
	}

	return nil
}
