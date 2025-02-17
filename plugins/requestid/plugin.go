package requestid

// https://pkg.go.dev/github.com/alx99/ika/plugins/requestid

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/alx99/ika"
	"github.com/alx99/ika/pluginutil"
	"github.com/google/uuid"
	"github.com/rs/xid"
	"github.com/segmentio/ksuid"
)

type plugin struct {
	next  ika.Handler
	cfg   pConfig
	genID func() (string, error)

	// once is used to ensure that printing of the xid info is only done once
	once sync.Once
}

func Factory() ika.PluginFactory {
	return &plugin{}
}

func (*plugin) Name() string {
	return "request-id"
}

func (f *plugin) New(ctx context.Context, ictx ika.InjectionContext, config map[string]any) (ika.Plugin, error) {
	p := &plugin{}

	if err := pluginutil.UnmarshalCfg(config, &p.cfg); err != nil {
		return nil, err
	}

	var err error
	p.genID, err = makeRandFun(p.cfg.Variant)
	if err != nil {
		return nil, err
	}

	if p.cfg.Variant == vXID {
		f.once.Do(func() {
			guid := xid.New()
			ictx.Logger.Log(ctx, slog.LevelInfo, "xid info", "pid", guid.Pid(), "machine", guid.Machine())
		})
	}

	return p, nil
}

func (p *plugin) Handler(next ika.Handler) ika.Handler {
	p.next = next
	return p
}

func (p *plugin) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
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

func (*plugin) Teardown(context.Context) error {
	return nil
}

func makeRandFun(variant string) (func() (string, error), error) {
	seed := [32]byte{}
	_, err := cryptoRand.Read(seed[:])
	if err != nil {
		return nil, err
	}

	switch variant {
	case vUUIDv4:
		rng := rand.NewChaCha8(seed)
		return func() (string, error) {
			uuid, err := uuid.NewRandomFromReader(rng)
			if err != nil {
				return "", err
			}
			return uuid.String(), nil
		}, nil
	case vUUIDv7:
		rng := rand.NewChaCha8(seed)
		return func() (string, error) {
			uuid, err := uuid.NewV7FromReader(rng)
			if err != nil {
				return "", err
			}
			return uuid.String(), nil
		}, nil
	case vKSUID:
		// Create a new random source for each plugin instance
		rng := rand.NewChaCha8(seed)
		return func() (string, error) {
			// Use a local buffer for random bytes
			var randBytes [16]byte
			_, err := rng.Read(randBytes[:])
			if err != nil {
				return "", err
			}
			// Create KSUID with current time and local random bytes
			kid, err := ksuid.FromParts(time.Now(), randBytes[:])
			if err != nil {
				return "", err
			}
			return kid.String(), nil
		}, nil
	case vXID:
		return func() (string, error) {
			return xid.NewWithTime(time.Now()).String(), nil
		}, nil
	}
	return nil, errors.New("unknown variant")
}

var (
	_ ika.OnRequestHook = &plugin{}
	_ ika.PluginFactory = &plugin{}
)
