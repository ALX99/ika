package ika

import (
	"flag"
	"fmt"
	"os"

	"github.com/alx99/ika/cmd/option"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/ika"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
)

func Run(opts ...option.Option) {
	flag.Parse()
	if *printVersion {
		fmt.Println("0.0.1")
		os.Exit(0)
	}

	cfg := config.Options{}
	for _, opt := range opts {
		if err := opt(&cfg); err != nil {
			fmt.Fprintf(os.Stderr, "failed to apply option: %s\n", err)
			os.Exit(1)
		}
	}

	ika.Run(*configPath, cfg)
}
