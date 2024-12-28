package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
)

type Method string

func (m *Method) UnmarshalJSON(data []byte) error {
	var tmp string
	if err := json.Unmarshal(data, &tmp); err != nil {
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
