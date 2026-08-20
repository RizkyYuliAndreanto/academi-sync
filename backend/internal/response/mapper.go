package response

import "net/http"

// StatusForCode maps an ErrorCode to its corresponding HTTP status code.
func StatusForCode(code ErrorCode) int {
	switch code {
	case CodeValidationError, CodeInvalidRequest:
		return http.StatusBadRequest
	case CodeUnauthenticated:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeResourceNotFound:
		return http.StatusNotFound
	case CodeMethodNotAllowed:
		return http.StatusMethodNotAllowed
	case CodeResourceConflict:
		return http.StatusConflict
	case CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case CodeUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
	case CodeRateLimited:
		return http.StatusTooManyRequests
	case CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case CodeInternalError:
		return http.StatusInternalServerError
	default:
		// Safe fallback to 500 Internal Server Error for unknown codes
		return http.StatusInternalServerError
	}
}
