package config

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"sigs.k8s.io/yaml"
)

type Config struct {
	Servers    []Server   `json:"servers"`
	Namespaces Namespaces `json:"namespaces"`
	Ika        Ika        `json:"ika"`
}

func Read(path string) (Config, error) {
	cfg := Config{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return cfg, err
	}

	if ext := filepath.Ext(path); ext == ".yaml" {
		data, err = yaml.YAMLToJSONStrict(data)
		if err != nil {
			return cfg, err
		}
	}

	defer f.Close()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}

	if len(cfg.Servers) < 1 {
		return cfg, errors.New("at least one server must be specified")
	}

	return cfg, nil
}

type Duration time.Duration

func (d Duration) LogValue() slog.Value {
	return slog.StringValue(time.Duration(d).String())
}

func (d Duration) Dur() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}
