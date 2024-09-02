package config

import (
	"fmt"
	"iter"
	"net/http"
	"os"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	Middlewares []Middleware
	Paths       map[string]Path
	Namespaces  map[string]Namespace
)

type Middleware struct {
	Name    string         `yaml:"name"`
	Enabled bool           `yaml:"enabled"`
	Args    map[string]any `yaml:",inline"`
}

// Names returns an iterator that yields the names of all enabled middlewares.
func (m Middlewares) Names() iter.Seq[string] {
	return func(yield func(string) bool) {
		for mw := range m.enabled() {
			if !yield(mw.Name) {
				return
			}
		}
	}
}

// enabled returns an iterator that yields all enabled middlewares.
func (m Middlewares) enabled() iter.Seq[Middleware] {
	return func(yield func(Middleware) bool) {
		for _, mw := range m {
			if mw.Enabled {
				if !yield(mw) {
					return
				}
			}
		}
	}
}

func (ns *Namespaces) UnmarshalYAML(value *yaml.Node) error {
	tmp := make(map[string]Namespace)

	if err := value.Decode(&tmp); err != nil {
		return err
	}

	for name, ns := range tmp {
		ns.Name = name
		tmp[name] = ns
	}

	*ns = tmp
	return nil
}

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
	Name                   string      `yaml:"name"`
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
	Host   string `yaml:"host"`
	Scheme string `yaml:"scheme"`
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
	RewritePath string      `yaml:"rewritePath"`
	Methods     []Method    `yaml:"methods"`
	Backends    []Backend   `yaml:"backends"`
	Middlewares Middlewares `yaml:"middlewares"`
}

// MergedMiddlewares returns an iterator that yields all enabled middlewares from the path and namespace.
func (p Path) MergedMiddlewares(ns Namespace) iter.Seq[Middleware] {
	return func(yield func(Middleware) bool) {
		for mw := range p.Middlewares.enabled() {
			if !yield(mw) {
				return
			}
		}

		for mw := range ns.Middlewares.enabled() {
			if !yield(mw) {
				return
			}
		}
	}
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
