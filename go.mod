module github.com/alx99/ika

go 1.24.0

require (
	// for colorful logs (only when IKA_DEBUG is set)
	github.com/lmittmann/tint v1.0.7
	// test dependencies
	github.com/matryer/is v1.4.1
	// for parsing YAML configuration
	sigs.k8s.io/yaml v1.4.0
)
