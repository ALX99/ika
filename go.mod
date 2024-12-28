module github.com/alx99/ika

go 1.23.0

require (
	// only used when IKA_DEBUG is set
	github.com/golang-cz/devslog v0.0.11
	// For parsing YAML configuration
	sigs.k8s.io/yaml v1.4.0
)

// test dependencies
require github.com/matryer/is v1.4.1
