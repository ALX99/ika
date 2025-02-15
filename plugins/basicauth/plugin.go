package basicauth

// https://pkg.go.dev/github.com/alx99/ika/plugins/basicauth

import (
	"context"
	"crypto/subtle"
	"errors"
	"net/http"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil"
	"github.com/alx99/ika/pluginutil/httperr"
)

type plugin struct {
	inUser, inPass   []byte
	outUser, outPass string
	next             ika.Handler
}

func Factory() ika.PluginFactory {
	return &plugin{}
}

func (*plugin) Name() string {
	return "basic-auth"
}

func (*plugin) New(ctx context.Context, ictx ika.InjectionContext, config map[string]any) (ika.Plugin, error) {
	p := &plugin{}

	cfg := pConfig{}
	if err := pluginutil.UnmarshalCfg(config, &cfg); err != nil {
		return nil, err
	}

	if cfg.Incoming != nil {
		inUser, inPass, err := cfg.Incoming.credentials()
		if err != nil {
			return nil, err
		}
		p.inUser, p.inPass = []byte(inUser), []byte(inPass)
	}

	if cfg.Outgoing != nil {
		var err error
		p.outUser, p.outPass, err = cfg.Outgoing.credentials()
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (p *plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if p.inUser != nil || p.inPass != nil {
		invalidCredsErr := httperr.New(http.StatusUnauthorized).
			WithErr(errors.New("invalid credentials")).
			WithTitle("Invalid credentials")

		user, pass, ok := r.BasicAuth()
		if !ok {
			return invalidCredsErr
		}

		if subtle.ConstantTimeCompare([]byte(user), []byte(p.inUser)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(p.inPass)) != 1 {
			return invalidCredsErr
		}
	}

	if p.outUser != "" || p.outPass != "" {
		r.SetBasicAuth(p.outUser, p.outPass)
	}

	return p.next.ServeHTTP(w, r)
}

func (*plugin) Teardown(context.Context) error {
	return nil
}

var (
	_ ika.Middleware    = &plugin{}
	_ ika.PluginFactory = &plugin{}
)
