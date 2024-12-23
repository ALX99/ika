package plugins

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"

	"github.com/alx99/ika/plugin"
)

type Dumper struct{}

func (Dumper) New(context.Context, plugin.InjectionContext) (plugin.Plugin, error) {
	return &Dumper{}, nil
}

func (Dumper) Name() string {
	return "dumper"
}

func (Dumper) Setup(context.Context, plugin.InjectionContext, map[string]any) error {
	return nil
}

func (Dumper) Teardown(context.Context) error { return nil }

func (a *Dumper) Handler(next plugin.ErrHandler) plugin.ErrHandler {
	log := slog.Default().With(slog.String("plugin", a.Name()))
	return plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		dumpedReq, err := httputil.DumpRequest(r, true)
		if err != nil {
			return err
		}

		recorder := httptest.NewRecorder()
		if err = next.ServeHTTP(multi{w, recorder}, r); err != nil {
			return fmt.Errorf("failed to serve request: %w", err)
		}

		dumpedResp, err := httputil.DumpResponse(recorder.Result(), true)
		if err != nil {
			return fmt.Errorf("failed to dump response: %w", err)
		}

		log.LogAttrs(r.Context(), slog.LevelDebug, "dumped request",
			slog.String("request", "\n"+string(dumpedReq)),
			slog.String("response", "\n"+string(dumpedResp)),
		)

		return nil
	})
}

type multi struct {
	http.ResponseWriter
	recorder *httptest.ResponseRecorder
}

func (m multi) WriteHeader(code int) {
	m.recorder.WriteHeader(code)
	m.ResponseWriter.WriteHeader(code)
}

func (m multi) Write(b []byte) (int, error) {
	maps.Copy(m.ResponseWriter.Header(), m.recorder.Header())
	if _, err := m.recorder.Write(b); err != nil {
		return 0, err
	}
	return m.ResponseWriter.Write(b)
}

func (m multi) Header() http.Header {
	return m.recorder.Header()
}
