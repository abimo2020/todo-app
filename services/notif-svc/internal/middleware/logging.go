package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{status: http.StatusOK, ResponseWriter: w}

		next.ServeHTTP(rec, r)

		slog.Info(
			"http-request",
			"path", r.URL.Path,
			"method", r.Method,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
