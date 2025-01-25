package requestid

import (
	"context"
	cryptoRand "crypto/rand"
	"errors"
	"math/rand/v2"
	"net/http"

	"github.com/alx99/ika"
	"github.com/google/uuid"
	"github.com/segmentio/ksuid"
)

type Plugin struct {
	cfg   config
	genID func() (string, error)
}

func (*Plugin) Name() string {
	return "request-id"
}

func (*Plugin) New(context.Context, ika.InjectionContext) (ika.Plugin, error) {
	return &Plugin{}, nil
}

func (p *Plugin) Setup(_ context.Context, _ ika.InjectionContext, config map[string]any) error {
	if err := toStruct(config, &p.cfg); err != nil {
		return err
	}

	if err := p.cfg.validate(); err != nil {
		return err
	}

	var err error
	p.genID, err = makeRandFun(p.cfg.Variant)
	return err
}

func (p *Plugin) ModifyRequest(r *http.Request) (*http.Request, error) {
	reqID, err := p.genID()
	if err != nil {
		return nil, err
	}

	if p.cfg.Override {
		r.Header.Set(p.cfg.Header, reqID)
	} else if p.cfg.Append {
		r.Header.Add(p.cfg.Header, reqID)
	} else {
		if r.Header.Get(p.cfg.Header) == "" {
			r.Header.Add(p.cfg.Header, reqID)
		}
	}

	return r, nil
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
			ksuid, err := ksuid.NewRandom()
			return ksuid.String(), err
		}, nil
	}
	return nil, errors.New("unknown variant")
}

var (
	_ ika.RequestModifier = &Plugin{}
	_ ika.PluginFactory   = &Plugin{}
)
