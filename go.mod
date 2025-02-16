module github.com/alx99/ika

go 1.24.0

require (
	github.com/lmittmann/tint v1.0.7 // for colorful logs
	github.com/matryer/is v1.4.1 // for testing
	github.com/mattn/go-isatty v0.0.20 // for checking if stdout is a TTY
	sigs.k8s.io/yaml v1.4.0 // for reading yaml files
)

require golang.org/x/sys v0.30.0 // indirect
