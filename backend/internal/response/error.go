package response

import "fmt"

// ErrorBody describes the JSON error payload.
type ErrorBody struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

// ErrorDetail is a type alias for ErrorBody for backward compatibility.
type ErrorDetail = ErrorBody

// AppError is an internal application error wrapping safe public info and hidden cause.
type AppError struct {
	Code          ErrorCode
	PublicMessage string
	Fields        map[string]string
	Cause         error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s (cause: %v)", e.Code, e.PublicMessage, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.PublicMessage)
}

// NewAppError creates a new AppError instance.
func NewAppError(code ErrorCode, publicMessage string, cause error) *AppError {
	return &AppError{
		Code:          code,
		PublicMessage: publicMessage,
		Cause:         cause,
	}
}
