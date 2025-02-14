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

	// Variant is the ID generation algorithm: UUIDv4, UUIDv7, KSUID, XID
	//
	// Defaults to "XID"
	Variant string `json:"variant"`

	// Override the existing header value if present
	//
	// Defaults to true
	Override *bool `json:"override"`

	// Append to the existing header value if present
	Append bool `json:"append"`

	// Expose controls whether to include the request ID in the response headers.
	// When enabled, the plugin will copy the final request ID to the response headers.
	//
	// The response header value follows these rules:
	// 1. If override=true: Uses the newly generated ID
	// 2. If append=true: Uses all request IDs (original and newly generated)
	// 3. If neither override nor append: Uses the existing ID if present, otherwise the new ID
	//
	// Defaults to true
	Expose *bool `json:"expose"`
}

const (
	vUUIDv4 = "UUIDv4"
	vUUIDv7 = "UUIDv7"
	vKSUID  = "KSUID"
	vXID    = "XID"
)

func (c *pConfig) SetDefaults() {
	c.Header = cmp.Or(c.Header, "X-Request-ID")
	c.Variant = cmp.Or(c.Variant, vXID)
	c.Override = cmp.Or(c.Override, &[]bool{true}[0])
	c.Expose = cmp.Or(c.Expose, &[]bool{true}[0])
}

func (c *pConfig) Validate() error {
	if c.Header == "" {
		return errors.New("header is required")
	}

	if !slices.Contains([]string{
		vUUIDv4,
		vUUIDv7,
		vKSUID,
		vXID,
	}, c.Variant) {
		return fmt.Errorf("invalid variant: %s", c.Variant)
	}

	if *c.Override && c.Append {
		return errors.New("override and append cannot both be true")
	}

	return nil
}
