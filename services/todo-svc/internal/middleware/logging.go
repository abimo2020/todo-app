package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		var event *zerolog.Event
		switch {
		case rec.status >= 500:
			event = log.Error()
		case rec.status >= 400:
			event = log.Warn()
		default:
			event = log.Info()
		}

		event.Str("path", r.URL.Path).
			Str("method", r.Method).
			Int("status", rec.status).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Msg("http request")
	})
}
