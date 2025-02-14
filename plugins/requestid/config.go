package requestid

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
)

type pConfig struct {
	// Header is the header to populate with the request ID
	//
	// Defaults to "X-Request-ID"
	Header string `json:"header"`

	// Variant is the ID generation algorithm: UUIDv4, UUIDv7, KSUID
	//
	// Defaults to "KSUID"
	Variant string `json:"variant"`

	// Override the existing header value if present
	//
	// Defaults to true
	Override *bool `json:"override"`

	// Append to the existing header value if present
	Append bool `json:"append"`
}

const (
	vUUIDv4 = "UUIDv4"
	vUUIDv7 = "UUIDv7"
	vKSUID  = "KSUID"
)

func (c *pConfig) SetDefaults() {
	c.Header = cmp.Or(c.Header, "X-Request-ID")
	c.Variant = cmp.Or(c.Variant, vKSUID)
	c.Override = cmp.Or(c.Override, &[]bool{true}[0])
}

func (c *pConfig) Validate() error {
	if c.Header == "" {
		return errors.New("header is required")
	}

	if !slices.Contains([]string{
		vUUIDv4,
		vUUIDv7,
		vKSUID,
	}, c.Variant) {
		return fmt.Errorf("invalid variant: %s", c.Variant)
	}

	if *c.Override && c.Append {
		return errors.New("override and append cannot both be true")
	}

	return nil
}
