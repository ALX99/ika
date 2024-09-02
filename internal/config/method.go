package config

import (
	"fmt"
	"net/http"
	"slices"

	"gopkg.in/yaml.v3"
)

type Method string

func (m *Method) UnmarshalYAML(value *yaml.Node) error {
	var tmp string
	if err := value.Decode(&tmp); err != nil {
		return err
	}

	validMethods := []string{
		http.MethodGet,
		http.MethodHead,
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodOptions,
		http.MethodTrace,
	}

	if !slices.Contains(validMethods, tmp) {
		return fmt.Errorf("invalid method: %s", tmp)
	}

	*m = Method(tmp)
	return nil
}
