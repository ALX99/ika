package basicauth

import (
	"cmp"
	"fmt"
	"os"
	"slices"
)

type pConfig struct {
	// Incoming is the configuration for incoming requests.
	Incoming *basicAuthConfig `json:"incoming"`

	// Outgoing is the configuration for outgoing requests.
	Outgoing *basicAuthConfig `json:"outgoing"`
}

type basicAuthConfig struct {
	// Type is how to look up the username and password.
	// The following types are supported: env
	// If omitted, username and password are used as-is.
	Type string `json:"type"`

	// Username is the username to use for basic auth.
	Username string `json:"username"`

	// Password is the password to use for basic auth.
	Password string `json:"password"`
}

func (c *pConfig) SetDefaults() {
	if c.Incoming != nil {
		c.Incoming.Type = cmp.Or(c.Incoming.Type, "static")
	}
	if c.Outgoing != nil {
		c.Outgoing.Type = cmp.Or(c.Outgoing.Type, "static")
	}
}

func (c *pConfig) Validate() error {
	if c.Incoming == nil && c.Outgoing == nil {
		return fmt.Errorf("at least one of incoming or outgoing must be set")
	}

	if c.Incoming != nil {
		if err := c.Incoming.validate(); err != nil {
			return fmt.Errorf("incoming validation failed: %w", err)
		}
	}
	if c.Outgoing != nil {
		if err := c.Outgoing.validate(); err != nil {
			return fmt.Errorf("outgoing validation failed: %w", err)
		}
	}
	return nil
}

func (c *basicAuthConfig) validate() error {
	if !slices.Contains([]string{"static", "env"}, c.Type) {
		return fmt.Errorf("type must be one of: static, env")
	}

	if c.Username == "" {
		return fmt.Errorf("username is required")
	}

	if c.Password == "" {
		return fmt.Errorf("password is required")
	}

	switch c.Type {
	case "static":
		return nil
	case "env":
		if _, ok := os.LookupEnv(c.Username); !ok {
			return fmt.Errorf("username environment variable not set")
		}
		if _, ok := os.LookupEnv(c.Password); !ok {
			return fmt.Errorf("password environment variable not set")
		}
	default:
		return fmt.Errorf("invalid type: %s", c.Type)
	}
	return nil
}

func (c *basicAuthConfig) credentials() (user, pass string, err error) {
	if c.Type == "static" {
		return c.Username, c.Password, nil
	}
	user, ok := os.LookupEnv(c.Username)
	if !ok {
		return "", "", fmt.Errorf("username environment variable not set")
	}
	pass, ok = os.LookupEnv(c.Password)
	if !ok {
		return "", "", fmt.Errorf("password environment variable not set")
	}
	return user, pass, nil
}
