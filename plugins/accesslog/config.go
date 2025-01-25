package accesslog

import (
	"encoding/json"
	"fmt"
)

type pConfig struct {
	// Headers contains the list of headers to log.
	Headers []string `json:"headers"`

	// IncludeRemoteAddr controls whether the remote address is included in the log.
	IncludeRemoteAddr bool `json:"logRemoteAddr"`
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
