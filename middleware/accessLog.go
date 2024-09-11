package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

func init() {
	err := Register("accessLog", Stateless(accessLog))
	if err != nil {
		panic(err)
	}
}

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		st := &statusRecorder{ResponseWriter: w}
		now := time.Now()
		next.ServeHTTP(st, r)
		slog.LogAttrs(r.Context(), slog.LevelInfo, "endpoint access",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("escapedPath", r.URL.EscapedPath()),
			slog.String("remote", r.RemoteAddr),
			slog.String("userAgent", r.UserAgent()),
			slog.Int("status", st.status),
			slog.Int64("duration", time.Since(now).Milliseconds()),
			slog.String("namespace", GetNamespace(r.Context())),
		)
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
