package gateway

import (
	"flag"
	"fmt"
	"os"

	"github.com/alx99/ika"

	"github.com/alx99/ika/internal/config"
	iika "github.com/alx99/ika/internal/ika"
)

var (
	printVersion = flag.Bool("version", false, "Print the version and exit.")
	configPath   = flag.String("config", "ika.yaml", "Path to the configuration file.")
)

// Run runs Ika gateway.
func Run(opts ...Option) {
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

	iika.Run(*configPath, cfg)
}

// Option represents an option for Run.
type Option func(*config.Options) error

// WithPlugin registers a plugin factory with Ika.
// Once registered, the plugin which the factory creates can be used in the configuration file.
func WithPlugin(factory ika.PluginFactory) Option {
	return func(cfg *config.Options) error {
		if cfg.Plugins == nil {
			cfg.Plugins = make(map[string]ika.PluginFactory)
		}
		if _, ok := cfg.Plugins[factory.Name()]; ok {
			return fmt.Errorf("plugin %q already registered", factory.Name())
		}
		cfg.Plugins[factory.Name()] = factory
		return nil
	}
}
