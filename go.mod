module github.com/alx99/ika

go 1.23.0

// dependencies
require (
	github.com/golang-cz/devslog v0.0.11 // only used when IKA_DEBUG is set
	github.com/valyala/bytebufferpool v1.0.0
	gopkg.in/yaml.v3 v3.0.1
)

// test dependencies
require github.com/matryer/is v1.4.1
