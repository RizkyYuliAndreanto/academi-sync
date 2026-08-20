package middleware

import (
	"log/slog"
	"time"

	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
)

// AccessLog creates a Gin middleware for structured access logging using log/slog.
func AccessLog(logger *slog.Logger, serviceName string, environment string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		duration := time.Since(start)
		durationMs := float64(duration.Microseconds()) / 1000.0

		reqID := response.GetRequestID(c)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if rawQuery != "" {
			path = path + "?" + rawQuery
		}

		attrs := []slog.Attr{
			slog.String("service", serviceName),
			slog.String("environment", environment),
			slog.String("request_id", reqID),
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Float64("duration_ms", durationMs),
			slog.String("client_ip", clientIP),
		}

		if errCode, exists := c.Get("error_code"); exists {
			if codeStr, ok := errCode.(string); ok && codeStr != "" {
				attrs = append(attrs, slog.String("error_code", codeStr))
			}
		}

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		logger.LogAttrs(c.Request.Context(), level, "http_request", attrs...)
	}
}
