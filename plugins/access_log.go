package plugins

import (
	"context"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/plugin"
)

var _ plugin.Middleware = &AccessLogger{}

type AccessLogger struct {
	pathPattern string
	namespace   string
}

func (AccessLogger) New(_ context.Context, _ plugin.InjectionContext) (plugin.Plugin, error) {
	return &AccessLogger{}, nil
}

func (AccessLogger) Name() string {
	return "accessLog"
}

func (a *AccessLogger) Setup(ctx context.Context, iCtx plugin.InjectionContext, config map[string]any) error {
	a.pathPattern = iCtx.PathPattern
	a.namespace = iCtx.Namespace
	return nil
}

func (AccessLogger) Teardown(context.Context) error { return nil }

func (a *AccessLogger) Handler(next plugin.ErrHandler) plugin.ErrHandler {
	return plugin.ErrHandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		now := time.Now()
		st := statusRecorder{ResponseWriter: w}
		err := next.ServeHTTP(&st, r)

		attrs := []slog.Attr{
			slog.String("method", r.Method),
			slog.String("path", request.GetPath(r)),
			slog.String("remote", r.RemoteAddr),
			slog.String("userAgent", r.UserAgent()),
			slog.Int64("status", st.status.Load()),
			slog.Int64("duration", time.Since(now).Milliseconds()),
			slog.String("namespace", a.namespace),
			slog.String("requestPattern", r.Pattern),
		}
		if a.pathPattern != "" {
			attrs = append(attrs, slog.String("pathPattern", a.pathPattern))
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		slog.LogAttrs(r.Context(), slog.LevelInfo, "endpoint access", attrs...)
		return err
	})
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
