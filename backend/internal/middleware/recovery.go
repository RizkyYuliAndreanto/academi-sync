package middleware

import (
	"log/slog"
	"runtime/debug"

	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
)

// Recovery middleware recovers from any panics and logs the stack trace safely.
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				reqID := response.GetRequestID(c)
				stackTrace := string(debug.Stack())

				logger.Error("panic recovered",
					slog.String("request_id", reqID),
					slog.Any("error", err),
					slog.String("stack", stackTrace),
				)

				c.Abort()

				if !c.Writer.Written() {
					response.Fail(c, response.CodeInternalError, "Terjadi kesalahan internal pada server")
				}
			}
		}()

		c.Next()
	}
}

// SecurityHeaders adds standard security headers to every HTTP response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
