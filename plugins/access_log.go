package plugins

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/plugin"
)

var _ plugin.Middleware = &AccessLogger{}

type AccessLogger struct {
	pathPattern string
	log         *slog.Logger
}

func (AccessLogger) New(_ context.Context, _ plugin.InjectionContext) (plugin.Plugin, error) {
	return &AccessLogger{}, nil
}

func (AccessLogger) Name() string {
	return "accessLog"
}

func (a *AccessLogger) Setup(ctx context.Context, iCtx plugin.InjectionContext, config map[string]any) error {
	a.pathPattern = iCtx.PathPattern
	a.log = iCtx.Logger
	return nil
}

func (AccessLogger) Teardown(context.Context) error { return nil }

func (a *AccessLogger) Handler(next plugin.Handler) plugin.Handler {
	return plugin.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		st := statusRecorder{ResponseWriter: w}
		now := time.Now()
		err := next.ServeHTTP(&st, r)
		end := time.Now()

		attrs := []slog.Attr{
			slog.Group("request",
				slog.String("method", r.Method),
				slog.String("path", request.GetPath(r)),
				slog.String("host", r.Host),
				slog.String("remoteAddr", r.RemoteAddr),
				slog.String("pattern", r.Pattern),
				slog.Any("cookies", cookies(r.Cookies()).LogValue()),
				slog.Any("headers", headers(r.Header).LogValue()),
			),
			slog.Group("response",
				slog.Int64("status", st.status.Load()),
				slog.Int64("duration", end.Sub(now).Milliseconds()),
			),
			slog.String("pathPattern", a.pathPattern),
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		a.log.LogAttrs(r.Context(), slog.LevelInfo, "endpoint access", attrs...)
		return err
	})
}

type cookies []*http.Cookie

var _ slog.LogValuer = cookies{}

func (c cookies) LogValue() slog.Value {
	if len(c) == 0 {
		return slog.Value{}
	}
	attrs := make([]slog.Attr, 0, len(c))
	for _, cookie := range c {
		attrs = append(attrs, slog.String(cookie.Name, cookie.Value))
	}
	return slog.GroupValue(attrs...)
}

type headers http.Header

var _ slog.LogValuer = headers{}

func (h headers) LogValue() slog.Value {
	if len(h) == 0 {
		return slog.Value{}
	}
	attrs := make([]slog.Attr, 0, len(h))
	for k, v := range h {
		attrs = append(attrs, slog.String(k, strings.Join(v, ", ")))
	}
	return slog.GroupValue(attrs...)
}

type statusRecorder struct {
	http.ResponseWriter
	status      atomic.Int64
	writeCalled atomic.Bool
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	if !w.writeCalled.Load() {
		w.status.Store(int64(statusCode))
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	w.writeCalled.Store(true)
	w.status.CompareAndSwap(0, http.StatusOK)
	return w.ResponseWriter.Write(b)
}
