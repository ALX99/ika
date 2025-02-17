// Package pluginutil provides utility functions for plugins.
package pluginutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"
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
// The config struct must be a pointer to a struct with JSON struct tags.
//
// The config struct can implement the [Defaulter] interface to set its default values.
// It can also implement the [Validator] interface to validate its values.
// Order of operations: UnmarshalCfg -> SetDefaults -> Validate
//
// This function supports unmarshaling string values into time.Duration (e.g. "1h", "30m")
func UnmarshalCfg(data map[string]any, config any) error {
	if config == nil {
		return errors.New("target is nil")
	}

	val := reflect.ValueOf(config)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("config must be a pointer to a struct")
	}

	// Convert input map to JSON bytes
	bs, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal config to JSON: %w", err)
	}

	// Try standard JSON unmarshal first - this handles most cases
	if err := json.Unmarshal(bs, config); err == nil {
		return applyDefaultsAndValidate(config)
	}

	// Parse JSON into a map for field-by-field processing
	var rawFields map[string]json.RawMessage
	if err := json.Unmarshal(bs, &rawFields); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	structVal := val.Elem()
	structType := structVal.Type()

	for i := range structType.NumField() {
		field := structVal.Field(i)
		jsonTag := structType.Field(i).Tag.Get("json")
		rawValue, ok := rawFields[jsonTag]

		// if the field is not settable, does not have a json tag, or is not in the rawFields, skip it
		if !field.CanSet() || jsonTag == "" || !ok {
			continue
		}

		if _, ok := field.Interface().(time.Duration); ok {
			var dur durAlias
			if err := json.Unmarshal(rawValue, &dur); err != nil {
				return fmt.Errorf("invalid duration for field %s: %w", jsonTag, err)
			}
			field.Set(reflect.ValueOf(time.Duration(dur)))
		}
	}

	return applyDefaultsAndValidate(config)
}

// applyDefaultsAndValidate handles the post-unmarshal operations
func applyDefaultsAndValidate(config any) error {
	// Set defaults if implemented
	if defaulter, ok := config.(Defaulter); ok {
		defaulter.SetDefaults()
	}

	// Validate if implemented
	if validator, ok := config.(Validator); ok {
		if err := validator.Validate(); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}

	return nil
}

// durAlias is a wrapper type that implements json.Unmarshaler for time.Duration
type durAlias time.Duration

func (d *durAlias) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = durAlias(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = durAlias(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}
