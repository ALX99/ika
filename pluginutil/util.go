// Package pluginutil provides utility functions for plugins.
package pluginutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// Error is a custom error type that implements interfaces for HTTP errors
// that is used by the default ika error handler.
type Error struct {
	title   string
	detail  string
	typeURI string
	status  int
}

// Error returns the detail of the error.
func (e *Error) Error() string {
	return e.detail
}

// Status returns the status code of the error.
func (e *Error) Status() int {
	return e.status
}

// TypeURI returns the type URI of the error.
func (e *Error) TypeURI() string {
	return e.typeURI
}

// Title returns the title of the error.
func (e *Error) Title() string {
	return e.title
}

// ToStruct unmarshals the given config map to the target struct.
//
// For this to work, the target struct must be a pointer to a struct
// with JSON struct tags.
func ToStruct(config map[string]any, target any) error {
	if target == nil {
		return errors.New("target is nil")
	}

	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
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
