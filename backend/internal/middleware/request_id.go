package middleware

import (
	"crypto/rand"
	"fmt"
	"regexp"

	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
)

const HeaderXRequestID = "X-Request-ID"

// validRequestIDRegex checks for valid alphanumeric, dash, and underscore characters up to 64 chars.
var validRequestIDRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_]{1,64}$`)

// RequestID middleware handles validation and generation of X-Request-ID headers.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		incomingID := c.GetHeader(HeaderXRequestID)

		var reqID string
		if incomingID != "" && validRequestIDRegex.MatchString(incomingID) {
			reqID = incomingID
		} else {
			reqID = generateUUIDv4()
		}

		c.Set(response.RequestIDKey, reqID)
		c.Header(HeaderXRequestID, reqID)

		c.Next()
	}
}

func generateUUIDv4() string {
	var uuid [16]byte
	_, err := rand.Read(uuid[:])
	if err != nil {
		// Fallback to timestamp based string if crypto/rand fails
		return fmt.Sprintf("fallback-%d", rand.Reader)
	}

	// Set version (4) and variant (RFC4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}
