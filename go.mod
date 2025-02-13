module github.com/alx99/ika

go 1.24.0

require (
	// for tests
	github.com/gkampitakis/go-snaps v0.5.10
	// for colorful logs (only when IKA_DEBUG is set)
	github.com/lmittmann/tint v1.0.7
	// for parsing YAML configuration
	sigs.k8s.io/yaml v1.4.0
)

// test dependencies
require (
	github.com/gkampitakis/ciinfo v0.3.1 // indirect
	github.com/gkampitakis/go-diff v1.3.2 // indirect
	github.com/goccy/go-yaml v1.15.13 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/maruel/natural v1.1.1 // indirect
	github.com/matryer/is v1.4.1
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/tidwall/gjson v1.18.0 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)
