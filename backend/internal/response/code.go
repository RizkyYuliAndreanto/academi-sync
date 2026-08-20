package response

// ErrorCode is a strong string type representing standard error codes.
type ErrorCode string

const (
	CodeValidationError      ErrorCode = "VALIDATION_ERROR"
	CodeInvalidRequest       ErrorCode = "INVALID_REQUEST"
	CodeUnauthenticated      ErrorCode = "UNAUTHENTICATED"
	CodeForbidden            ErrorCode = "FORBIDDEN"
	CodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	CodeResourceConflict     ErrorCode = "RESOURCE_CONFLICT"
	CodeRateLimited          ErrorCode = "RATE_LIMITED"
	CodeInternalError        ErrorCode = "INTERNAL_ERROR"
	CodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"
	CodeMethodNotAllowed     ErrorCode = "METHOD_NOT_ALLOWED"
	CodePayloadTooLarge      ErrorCode = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType ErrorCode = "UNSUPPORTED_MEDIA_TYPE"
)
