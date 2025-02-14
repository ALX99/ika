// Package pluginutil provides utility functions for plugins.
package pluginutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Defaulter is an interface that can be implemented by a struct to set its default values.
type Defaulter interface {
	SetDefaults()
}

// Validator is an interface that can be implemented by a struct to validate its values.
type Validator interface {
	Validate() error
}

// UnmarshalCfg unmarshals the given config map to the target struct.
// The config struct must be a pointer to a struct
// with JSON struct tags.
//
// The config struct can implement the [Defaulter] interface to set its default values.
// It can also implement the [Validator] interface to validate its values.
// Order of operations: UnmarshalCfg -> SetDefaults -> Validate
func UnmarshalCfg(data map[string]any, config any) error {
	if config == nil {
		return errors.New("target is nil")
	}

	if reflect.ValueOf(config).Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	bs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	err = json.Unmarshal(bs, config)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON to config: %w", err)
	}

	// Set the default values
	if defaulter, ok := config.(Defaulter); ok {
		defaulter.SetDefaults()
	}

	// Validate the config
	if validator, ok := config.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}
