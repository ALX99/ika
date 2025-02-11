package basicauth

import (
	"context"
	"crypto/subtle"
	"net/http"
	"net/url"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil"
)

type Plugin struct {
	inUser, inPass   []byte
	outUser, outPass string
	inEncoding       string
	next             ika.Handler
}

func (*Plugin) New(context.Context, ika.InjectionContext) (ika.Plugin, error) {
	return &Plugin{}, nil
}

func (*Plugin) Name() string {
	return "basic-auth"
}

func (p *Plugin) Setup(_ context.Context, _ ika.InjectionContext, config map[string]any) error {
	cfg := pConfig{}
	if err := pluginutil.ToStruct(config, &cfg); err != nil {
		return err
	}

	if err := cfg.validate(); err != nil {
		return err
	}

	if cfg.Incoming != nil {
		inUser, inPass, err := cfg.Incoming.credentials()
		if err != nil {
			return err
		}
		p.inUser, p.inPass = []byte(inUser), []byte(inPass)
	}

	if cfg.Outgoing != nil {
		var err error
		p.outUser, p.outPass, err = cfg.Outgoing.credentials()
		if err != nil {
			return err
		}
	}
	p.inEncoding = cfg.Incoming.Encoding

	return nil
}

func (p *Plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if p.inUser != nil || p.inPass != nil {
		user, pass, ok := r.BasicAuth()
		if !ok {
			return pluginutil.NewError(http.StatusText(http.StatusUnauthorized), "invalid credentials", "", http.StatusUnauthorized)
		}

		if p.inEncoding == "urlencoding" {
			var err error
			user, err = url.QueryUnescape(user)
			if err != nil {
				return err
			}
			pass, err = url.QueryUnescape(pass)
			if err != nil {
				return err
			}
		}

		if subtle.ConstantTimeCompare([]byte(user), []byte(p.inUser)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(p.inPass)) != 1 {
			return pluginutil.NewError(http.StatusText(http.StatusUnauthorized), "invalid credentials", "", http.StatusUnauthorized)
		}
	}

	if p.outUser != "" || p.outPass != "" {
		r.SetBasicAuth(p.outUser, p.outPass)
	}

	return p.next.ServeHTTP(w, r)
}

func (*Plugin) Teardown(context.Context) error {
	return nil
}

var (
	_ ika.Middleware    = &Plugin{}
	_ ika.PluginFactory = &Plugin{}
)
