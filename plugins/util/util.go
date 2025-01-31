// Package util provides utility functions for plugins.
// It is not a plugin itself.
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// ToStruct unmarshals the given config map to the target struct.
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
