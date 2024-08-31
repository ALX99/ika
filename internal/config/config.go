package config

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Middlewares map[string]map[string]any
	Namespaces  map[string]Namespace
	Paths       map[string]Path
)

type Ika struct {
	Server                  Server        `yaml:"server"`
	Namespaces              Namespaces    `yaml:"namespaces"`
	GracefulShutdownTimeout time.Duration `yaml:"gracefulShutdownTimeout"`
}

type Server struct {
	Addr                         string        `yaml:"addr"`
	DisableGeneralOptionsHandler bool          `yaml:"disableGeneralOptionsHandler"`
	ReadTimeout                  time.Duration `yaml:"readTimeout"`
	ReadHeaderTimeout            time.Duration `yaml:"readHeaderTimeout"`
	WriteTimeout                 time.Duration `yaml:"writeTimeout"`
	IdleTimeout                  time.Duration `yaml:"idleTimeout"`
	MaxHeaderBytes               int           `yaml:"maxHeaderBytes"`
}

type Namespace struct {
	Backends               []Backend   `yaml:"backends"`
	Transport              Transport   `yaml:"transport"`
	Paths                  Paths       `yaml:"paths"`
	Middlewares            Middlewares `yaml:"middlewares"`
	Hosts                  []string    `yaml:"hosts"`
	DisableNamespacedPaths bool        `yaml:"disableNamespacedPaths"`
}

func (ns *Namespace) UnmarshalYAML(value *yaml.Node) error {
	type alias Namespace
	tmp := alias{}

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	if len(tmp.Hosts) == 0 && tmp.DisableNamespacedPaths {
		return fmt.Errorf("namespace has no hosts and namespaced paths are disabled")
	}

	*ns = Namespace(tmp)
	return nil
}

type Backend struct {
	Host        string `yaml:"host"`
	Scheme      string `yaml:"scheme"`
	RewritePath string `yaml:"rewritePath"`
}

type Transport struct {
	DisableKeepAlives      bool          `yaml:"disableKeepAlives"`
	DisableCompression     bool          `yaml:"disableCompression"`
	MaxIdleConns           int           `yaml:"maxIdleConns"`
	MaxIdleConnsPerHost    int           `yaml:"maxIdleConnsPerHost"`
	MaxConnsPerHost        int           `yaml:"maxConnsPerHost"`
	IdleConnTimeout        time.Duration `yaml:"idleConnTimeout"`
	ResponseHeaderTimeout  time.Duration `yaml:"responseHeaderTimeout"`
	ExpectContinueTimeout  time.Duration `yaml:"expectContinueTimeout"`
	MaxResponseHeaderBytes int64         `yaml:"maxResponseHeaderBytes"`
	WriteBufferSize        int           `yaml:"writeBufferSize"`
	ReadBufferSize         int           `yaml:"readBufferSize"`
}

type Path struct {
	Methods     []Method    `yaml:"methods"`
	Backends    []Backend   `yaml:"backends"`
	Middlewares Middlewares `yaml:"middlewares"`
}

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

func ReadConfig(path string) (Ika, error) {
	cfg := Ika{}

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	return cfg, yaml.NewDecoder(f).Decode(&cfg)
}
