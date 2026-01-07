package middleware

import (
	"log/slog"
	"math"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func GinLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := time.Now()

		defer func() {
			t2 := time.Now()
			status := c.Writer.Status()
			bytes := int(math.Max(float64(c.Writer.Size()), 0))
			values := []any{
				slog.String("log_type", "http_request"),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.String("remote", c.Request.RemoteAddr),
				slog.Int("status", status),
				slog.Int("bytes", bytes),
				slog.Duration("duration", t2.Sub(t1)),
				slog.String("request_id", requestid.Get(c)),
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

		c.Next()
	}
}
