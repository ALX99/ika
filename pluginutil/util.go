// Package pluginutil provides utility functions for plugins.
package pluginutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ToStruct unmarshals the given config map to the target struct.
//
// For this to work, the target struct must be a pointer to a struct
// with JSON struct tags.
func ToStruct(config map[string]any, target any) error {
	if target == nil {
		return errors.New("target is nil")
	}

	if reflect.ValueOf(target).Kind() == reflect.Ptr {
		target = reflect.ValueOf(target).Elem().Interface()
	}

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
