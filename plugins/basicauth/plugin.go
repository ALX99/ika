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
	inCreds          []credential
	outUser, outPass string
	next             ika.Handler
}

type credential struct {
	user []byte
	pass []byte
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
		p.inCreds = make([]credential, len(cfg.Incoming.Credentials))
		for i, cred := range cfg.Incoming.Credentials {
			user, pass, err := cred.credentials()
			if err != nil {
				return nil, err
			}
			p.inCreds[i] = credential{
				user: []byte(user),
				pass: []byte(pass),
			}
		}
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
	if len(p.inCreds) > 0 {
		invalidCredsErr := httperr.New(http.StatusUnauthorized).
			WithErr(errors.New("invalid credentials")).
			WithTitle("Invalid credentials")

		user, pass, ok := r.BasicAuth()
		if !ok {
			return invalidCredsErr
		}

		userBytes := []byte(user)
		passBytes := []byte(pass)

		for _, cred := range p.inCreds {
			if subtle.ConstantTimeCompare(userBytes, cred.user) == 1 &&
				subtle.ConstantTimeCompare(passBytes, cred.pass) == 1 {
				// Found valid credentials
				goto authorized
			}
		}
		return invalidCredsErr
	}
authorized:
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
