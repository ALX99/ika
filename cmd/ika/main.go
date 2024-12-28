package main

import (
	"github.com/alx99/ika"
	"github.com/alx99/ika/cmd/option"
)

var opts []option.Option

func main() {
	ika.Run(opts...)
}
