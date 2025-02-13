package requestid

// https://pkg.go.dev/github.com/alx99/ika/plugins/requestid

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

type Plugin struct {
	next  ika.Handler
	cfg   pConfig
	genID func() (string, error)
}

func (*Plugin) Name() string {
	return "request-id"
}

func (*Plugin) New(ctx context.Context, ictx ika.InjectionContext, config map[string]any) (ika.Plugin, error) {
	p := &Plugin{}

	if err := pluginutil.UnmarshalCfg(config, &p.cfg); err != nil {
		return nil, err
	}

	var err error
	p.genID, err = makeRandFun(p.cfg.Variant)
	if err != nil {
		return nil, err
	}

	if p.cfg.Variant == vXID {
		guid := xid.New()
		ictx.Logger.Log(ctx, slog.LevelInfo, "xid info", "pid", guid.Pid(), "machine", guid.Machine())
	}

	return p, nil
}

func (p *Plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *Plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	reqID, err := p.genID()
	if err != nil {
		return err
	}

	existingID := r.Header.Get(p.cfg.Header)

	switch {
	case *p.cfg.Override:
		r.Header.Set(p.cfg.Header, reqID)
	case p.cfg.Append:
		r.Header.Add(p.cfg.Header, reqID)
	case existingID == "":
		r.Header.Add(p.cfg.Header, reqID)
	}

	if *p.cfg.Expose {
		switch {
		case existingID == "" || *p.cfg.Override:
			w.Header().Set(p.cfg.Header, reqID)
		case p.cfg.Append:
			w.Header().Add(p.cfg.Header, r.Header.Get(p.cfg.Header))
		default:
			w.Header().Set(p.cfg.Header, existingID)
		}
	}

	return p.next.ServeHTTP(w, r)
}

func (*Plugin) Teardown(context.Context) error {
	return nil
}

func makeRandFun(variant string) (func() (string, error), error) {
	seed := [32]byte{}
	_, err := cryptoRand.Read(seed[:])
	if err != nil {
		return nil, err
	}
	chacha := rand.NewChaCha8(seed)

	switch variant {
	case vUUIDv4:
		uuid.SetRand(chacha)
		uuid.EnableRandPool()
		return func() (string, error) {
			uuid, err := uuid.NewRandom()
			if err != nil {
				return "", err
			}
			return uuid.String(), nil
		}, nil
	case vUUIDv7:
		uuid.SetRand(chacha)
		uuid.EnableRandPool()
		return func() (string, error) {
			uuid, err := uuid.NewV7()
			if err != nil {
				return "", err
			}
			return uuid.String(), nil
		}, nil
	case vKSUID:
		ksuid.SetRand(chacha)
		return func() (string, error) {
			ksuid, err := ksuid.NewRandomWithTime(time.Now())
			return ksuid.String(), err
		}, nil
	case vXID:
		return func() (string, error) {
			return xid.NewWithTime(time.Now()).String(), nil
		}, nil
	}
	return nil, errors.New("unknown variant")
}

var (
	_ ika.OnRequestHook = &Plugin{}
	_ ika.PluginFactory = &Plugin{}
)
