package requestid

import (
	"context"
	"crypto/rand"
	"net/http"

	"github.com/alx99/ika"
	"github.com/google/uuid"
)

type requestID struct {
	cfg config
}

func (p *requestID) ModifyRequest(r *http.Request) (*http.Request, error) {
	var reqID string

	if p.cfg.Variant == uuidV7 {
		uuid, err := uuid.NewV7()
		if err != nil {
			return nil, err // as of Go 1.24 this should never happen
		}
		reqID = uuid.String()
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

func (*requestID) Name() string {
	return "requestID"
}

func (*requestID) New(context.Context, ika.InjectionContext) (ika.Plugin, error) {
	return &requestID{}, nil
}

func (p *requestID) Setup(_ context.Context, _ ika.InjectionContext, config map[string]any) error {
	uuid.SetRand(rand.Reader)
	uuid.EnableRandPool()

	if err := toStruct(config, &p.cfg); err != nil {
		return err
	}
	return p.cfg.validate()
}

func (*requestID) Teardown(context.Context) error {
	return nil
}

var (
	_ ika.RequestModifier = &requestID{}
	_ ika.PluginFactory   = &requestID{}
)
