package requestid

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
)

type config struct {
	// Header is the header to populate with the request ID.
	Header string `json:"header"`

	// Override if true, will override the request id header if it already exists.
	Override bool `json:"override"`

	// Append if true, will append the request id header if it already exists.
	Append bool `json:"append"`

	// Variant is the request id variant to generate.
	// The following variants are supported: UUIDv4, UUIDv7, KSUID
	Variant string `json:"variant"`
}

const (
	vUUIDv4 = "UUIDv4"
	vUUIDv7 = "UUIDv7"
	vKSUID  = "KSUID"
)

func (c *config) validate() error {
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

func toStruct(config map[string]any, target any) error {
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to config: %w", err)
	}

	return nil
}
