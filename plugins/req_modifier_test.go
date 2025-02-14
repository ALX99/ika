package plugins

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func BenchmarkRewritePath(b *testing.B) {
	is := is.New(b)
	config := map[string]any{
		"path": "/new/{path}",
	}
	iCtx := ika.InjectionContext{
		Route:  "/old/{path}",
		Logger: slog.New(slog.DiscardHandler),
	}
	p, err := (&ReqModifier{}).New(b.Context(), iCtx, config)
	if err != nil {
		b.Fatal(err)
	}
	rm := p.(*ReqModifier)

	rm.setupPathRewrite(iCtx.Route)

	req, _ := http.NewRequest("GET", "http://example.com/old/test", nil)

	for b.Loop() {
		req.URL.Path = "/old/test"
		err := rm.rewritePath(req)
		is.NoErr(err)

		if req.URL.Path != "/new/test" {
			b.Fatalf("unexpected path: %s", req.URL.Path)
		}
	}
}
