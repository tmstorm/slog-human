package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func ChiLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			defer func() {
				t2 := time.Now()
				status := ww.Status()
				bytes := ww.BytesWritten()
				values := []any{
					slog.String("log_type", "http_request"),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote", r.RemoteAddr),
					slog.Int("status", status),
					slog.Int("bytes", bytes),
					slog.Duration("duration", t2.Sub(t1)),
					slog.String("request_id", middleware.GetReqID(r.Context())),
				}
				switch {
				case status >= 200 && status <= 299:
					logger.Info("", values...)
				case status >= 300 && status <= 499:
					logger.Warn("", values...)
				case status >= 500 && status <= 599:
					logger.Error("", values...)
				default:
					logger.Info("", values...)
				}
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
