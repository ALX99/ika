package config

type Backend struct {
	Host   string `yaml:"host"`
	Scheme string `yaml:"scheme"`
}
