package requestid

import (
	"errors"
	"slices"
)

type pConfig struct {
	// Header is the header to populate with the request ID
	Header string `json:"header"`

	// Variant is the ID generation algorithm: UUIDv4, UUIDv7, KSUID
	Variant string `json:"variant"`

	// Override the existing header value if present
	Override bool `json:"override"`

	// Append to the existing header value if present
	Append bool `json:"append"`
}

const (
	vUUIDv4 = "UUIDv4"
	vUUIDv7 = "UUIDv7"
	vKSUID  = "KSUID"
)

func (c *pConfig) validate() error {
	if c.Header == "" {
		return errors.New("header is required")
	}

	if !slices.Contains([]string{
		vUUIDv4,
		vUUIDv7,
		vKSUID,
	}, c.Variant) {
		return errors.New("invalid variant")
	}

	return nil
}
