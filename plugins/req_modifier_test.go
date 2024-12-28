package plugins

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"testing"

	"github.com/alx99/ika/plugin"
)

func BenchmarkRewritePath(b *testing.B) {
	p, err := ReqModifier{}.New(context.Background(), plugin.InjectionContext{})
	if err != nil {
		b.Fatal(err)
	}
	rm := p.(*ReqModifier)

	config := map[string]any{
		"path": "/new/{path}",
	}
	iCtx := plugin.InjectionContext{
		PathPattern: "/old/{path}",
		Logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
	rm.Setup(context.Background(), iCtx, config)
	rm.setupPathRewrite(iCtx.PathPattern)

	req, _ := http.NewRequest("GET", "http://example.com/old/test", nil)

	for n := 0; n < b.N; n++ {
		req.URL.Path = "/old/test"
		rm.rewritePath(req)
		if req.URL.Path != "/new/test" {
			b.Fatalf("unexpected path: %s", req.URL.Path)
		}
	}
}
