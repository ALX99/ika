package fail2ban

import (
	"cmp"
	"errors"
	"time"
)

type pConfig struct {
	// MaxAttempts is the number of failed attempts before banning
	MaxAttempts uint64 `json:"maxAttempts"`

	// Window is the time window to track failed attempts
	Window time.Duration `json:"window"`

	// BanDuration is how long to ban IPs that exceed MaxAttempts
	//
	// Defaults to `window` * 2
	BanDuration time.Duration `json:"banDuration"`

	// IDHeader is the header containing the identifier to ban.
	// If empty, the remote IP address will be used.
	// Common values might be "X-Real-IP", "X-Forwarded-For", or "CF-Connecting-IP"
	IDHeader string `json:"idHeader"`
}

func (c *pConfig) SetDefaults() {
	c.BanDuration = cmp.Or(c.BanDuration, c.Window*2)
}

func (c *pConfig) Validate() error {
	if c.MaxAttempts <= 0 {
		return errors.New("maxAttempts must be greater than 0")
	}
	if c.Window <= 0 {
		return errors.New("window must be greater than 0")
	}
	if c.BanDuration <= 0 {
		return errors.New("banDuration must be greater than 0")
	}
	return nil
}
