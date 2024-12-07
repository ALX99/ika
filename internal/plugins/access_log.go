package plugins

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/alx99/ika/internal/request"
	"github.com/alx99/ika/plugin"
)

var (
	_ plugin.Plugin     = &AccessLogger{}
	_ plugin.Middleware = &AccessLogger{}
)

type AccessLogger struct {
	pathPattern string
	namespace   string
}

func (AccessLogger) New(context.Context) (plugin.Plugin, error) {
	return &AccessLogger{}, nil
}

func (AccessLogger) Name() string {
	return "accessLog"
}

func (AccessLogger) Capabilities() []plugin.Capability {
	return []plugin.Capability{plugin.CapMiddleware}
}

func (AccessLogger) InjectionLevels() []plugin.InjectionLevel {
	return []plugin.InjectionLevel{plugin.LevelPath}
}

func (a *AccessLogger) Setup(ctx context.Context, context plugin.InjectionContext, config map[string]any) error {
	a.pathPattern = context.PathPattern
	a.namespace = context.Namespace
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
			slog.String("pathPattern", a.pathPattern),
			slog.String("remote", r.RemoteAddr),
			slog.String("userAgent", r.UserAgent()),
			slog.Int("status", st.status),
			slog.Int64("duration", time.Since(now).Milliseconds()),
			slog.String("namespace", a.namespace),
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
	status int
}

func (w *statusRecorder) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}
