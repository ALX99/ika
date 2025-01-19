package main

import (
	"github.com/alx99/ika/gateway"
)

var opts []gateway.Option

func main() {
	gateway.Run(opts...)
}
