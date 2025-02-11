package plugins

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/alx99/ika"
	"github.com/matryer/is"
)

func BenchmarkRewritePath(b *testing.B) {
	is := is.New(b)
	p, err := (&ReqModifier{}).New(context.Background(), ika.InjectionContext{})
	if err != nil {
		b.Fatal(err)
	}
	rm := p.(*ReqModifier)

	config := map[string]any{
		"path": "/new/{path}",
	}
	iCtx := ika.InjectionContext{
		PathPattern: "/old/{path}",
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	err = rm.Setup(context.Background(), iCtx, config)
	is.NoErr(err)

	rm.setupPathRewrite(iCtx.PathPattern)

	req, _ := http.NewRequest("GET", "http://example.com/old/test", nil)

	for n := 0; n < b.N; n++ {
		req.URL.Path = "/old/test"
		err := rm.rewritePath(req)
		is.NoErr(err)

		if req.URL.Path != "/new/test" {
			b.Fatalf("unexpected path: %s", req.URL.Path)
		}
	}
}
