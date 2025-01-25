package requestid

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"slices"
	"strings"
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

	// Encoding is the encoding that is used for the username and password.
	// For outgoing requests, the encoding is used to encode the username and password.
	// For incoming requests, the encoding is used to decode the username and password.
	// The following encodings are supported: urlencoding
	// If omitted, no encoding is used.
	Encoding string `json:"encoding"`

	// Username is the username to use for basic auth.
	Username string `json:"username"`

	// Password is the password to use for basic auth.
	Password string `json:"password"`
}

func (c *pConfig) validate() error {
	if c.Incoming == nil && c.Outgoing == nil {
		return fmt.Errorf("at least one of incoming or outgoing must be set")
	}

	if c.Incoming != nil {
		if err := c.Incoming.validate(); err != nil {
			return fmt.Errorf("incoming: %w", err)
		}
	}
	if c.Outgoing != nil {
		if err := c.Outgoing.validate(); err != nil {
			return fmt.Errorf("outgoing: %w", err)
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
	if c.Type == "env" {
		if _, ok := os.LookupEnv(c.Username); !ok {
			return fmt.Errorf("username environment variable not set")
		}
		if _, ok := os.LookupEnv(c.Password); !ok {
			return fmt.Errorf("password environment variable not set")
		}
	}
	if c.Encoding != "" && !slices.Contains([]string{"urlencoding"}, c.Encoding) {
		return fmt.Errorf("encoding must be one of: urlencoding")
	}
	if c.Encoding == "" && strings.Contains(c.Username, ":") {
		return fmt.Errorf("username contains a colon, encoding must be set")
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
	if c.Encoding == "urlencoding" {
		user, err = url.QueryUnescape(user)
		if err != nil {
			return "", "", fmt.Errorf("failed to unescape username: %w", err)
		}
		pass, err = url.QueryUnescape(pass)
		if err != nil {
			return "", "", fmt.Errorf("failed to unescape password: %w", err)
		}
	}
	return user, pass, nil
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
