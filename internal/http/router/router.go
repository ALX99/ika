package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/internal/teardown"
)

type Router struct {
	tder teardown.Teardowner
	mux  *http.ServeMux
	cfg  config.Config
	opts config.ComptimeOpts
	log  *slog.Logger
}

func New(cfg config.Config, opts config.ComptimeOpts, log *slog.Logger) (*Router, error) {
	return &Router{
		tder: make(teardown.Teardowner, 0),
		mux:  http.NewServeMux(),
		cfg:  cfg,
		opts: opts,
		log:  log,
	}, nil
}

func (r *Router) Build(ctx context.Context) error {
	r.log.Info("Building router", "namespaceCount", len(r.cfg.Namespaces))

	for nsName, ns := range r.cfg.Namespaces {
		now := time.Now()
		builder, err := newNSBuilder(ctx, r.mux, nsName, ns, r.log, r.opts.Plugins)
		if err != nil {
			return err
		}

		if err := builder.build(ctx); err != nil {
			return err
		}
		r.tder.Add(builder.teardown)
		r.log.Debug("Built namespace", "ns", nsName, "dur", time.Since(now))
	}

	return nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Shutdown shuts down the router
func (r *Router) Shutdown(ctx context.Context) error {
	return r.tder.Teardown(ctx)
}
