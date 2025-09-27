package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// NewStructuredLogger is a custom middleware that provides structured logging for requests.
func NewStructuredLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			tww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t_start := time.Now()
			defer func() {
				status := tww.Status()
				latency := time.Since(t_start)

				requestAttrs := slog.Group("request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
				)

				responseAttrs := slog.Group("response",
					slog.Int("status", status),
					slog.Int("bytes", tww.BytesWritten()),
					slog.String("latency", latency.String()),
				)

				if status >= 500 {
					logger.Error("server error", requestAttrs, responseAttrs)
				} else {
					logger.Info("request completed", requestAttrs, responseAttrs)
				}
			}()

			next.ServeHTTP(tww, r)
		}
		return http.HandlerFunc(fn)
	}
}
