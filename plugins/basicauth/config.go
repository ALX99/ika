package basicauth

import (
	"cmp"
	"fmt"
	"os"
	"slices"
)

type pConfig struct {
	// Incoming is the configuration for incoming requests.
	Incoming *incomingConfig `json:"incoming"`

	// Outgoing is the configuration for outgoing requests.
	Outgoing *basicAuthConfig `json:"outgoing"`
}

type incomingConfig struct {
	// Credentials is a list of valid credentials for incoming requests
	Credentials []namedCredential `json:"credentials"`

	// Strip determines whether to remove the basic auth credentials
	// from the request after successful authentication
	Strip bool `json:"strip"`
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

type namedCredential struct {
	// Name is a unique identifier for this credential
	Name string `json:"name"`

	basicAuthConfig
}

func (c *pConfig) SetDefaults() {
	if c.Incoming != nil {
		for i := range c.Incoming.Credentials {
			c.Incoming.Credentials[i].Type = cmp.Or(c.Incoming.Credentials[i].Type, "static")
		}
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

func (c *incomingConfig) validate() error {
	if len(c.Credentials) == 0 {
		return fmt.Errorf("at least one credential must be provided")
	}

	names := make(map[string]bool)
	for _, cred := range c.Credentials {
		if cred.Name == "" {
			return fmt.Errorf("credential name is required")
		}
		if names[cred.Name] {
			return fmt.Errorf("duplicate credential name: %s", cred.Name)
		}
		names[cred.Name] = true

		if err := cred.validate(); err != nil {
			return fmt.Errorf("credential %s validation failed: %w", cred.Name, err)
		}
	}
	return nil
}

func (c *namedCredential) validate() error {
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

func (c *namedCredential) credentials() (user, pass string, err error) {
	return c.basicAuthConfig.credentials()
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
