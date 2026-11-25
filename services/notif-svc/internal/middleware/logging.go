package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
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

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()

		var event *zerolog.Event
		switch {
		case status >= 500:
			event = log.Error()
		case status >= 400:
			event = log.Warn()
		default:
			event = log.Info()
		}
		event.Str("path", c.Request.URL.Path).
			Str("method", c.Request.Method).
			Str("route", c.FullPath()).
			Int("status", status).
			Int64("duration_ms", time.Since(start).Milliseconds()).
			Msg("http-request")
	}
}
