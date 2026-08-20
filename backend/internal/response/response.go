package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

// SuccessEnvelope is the distinct success response structure.
type SuccessEnvelope struct {
	Data      any            `json:"data"`
	Meta      map[string]any `json:"meta,omitempty"`
	RequestID string         `json:"request_id"`
}

// ErrorEnvelope is the distinct error response structure.
type ErrorEnvelope struct {
	Error     ErrorBody `json:"error"`
	RequestID string    `json:"request_id"`
}

// GetRequestID extracts the request_id from gin context or response header fallback.
func GetRequestID(c *gin.Context) string {
	if val, exists := c.Get(RequestIDKey); exists {
		if idStr, ok := val.(string); ok && idStr != "" {
			return idStr
		}
	}
	if headerID := c.Writer.Header().Get("X-Request-ID"); headerID != "" {
		return headerID
	}
	return ""
}

// OK sends a 200 OK success response envelope.
func OK(c *gin.Context, data any) {
	Success(c, http.StatusOK, data, nil)
}

// Created sends a 201 Created success response envelope.
func Created(c *gin.Context, data any) {
	Success(c, http.StatusCreated, data, nil)
}

// Success sends a success response envelope with custom HTTP status code and optional meta.
func Success(c *gin.Context, status int, data any, meta map[string]any) {
	reqID := GetRequestID(c)
	c.JSON(status, SuccessEnvelope{
		Data:      data,
		Meta:      meta,
		RequestID: reqID,
	})
}

// Fail sends an error response envelope mapped to the proper HTTP status.
func Fail(c *gin.Context, code ErrorCode, message string) {
	failWithFields(c, code, message, nil)
}

// ValidationFail sends a 400 VALIDATION_ERROR response envelope with field validation details.
func ValidationFail(c *gin.Context, message string, fields map[string]string) {
	failWithFields(c, CodeValidationError, message, fields)
}

// AppErr handles an AppError instance by logging cause internally and returning safe public error to client.
func AppErr(c *gin.Context, err *AppError) {
	if err == nil {
		InternalServerError(c)
		return
	}
	failWithFields(c, err.Code, err.PublicMessage, err.Fields)
}

// NoContent sends a 204 No Content response with no body.
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
	c.Writer.WriteHeaderNow()
}

// failWithFields is the internal implementation for error responses.
func failWithFields(c *gin.Context, code ErrorCode, message string, fields map[string]string) {
	status := StatusForCode(code)
	reqID := GetRequestID(c)

	c.Set("error_code", string(code))

	c.JSON(status, ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Fields:  fields,
		},
		RequestID: reqID,
	})
}

// Backward-compatibility wrappers

// JSON sends a standard success response envelope (legacy helper).
func JSON(c *gin.Context, statusCode int, data any, meta any) {
	var metaMap map[string]any
	if m, ok := meta.(map[string]any); ok {
		metaMap = m
	}
	Success(c, statusCode, data, metaMap)
}

// Error sends a standard error response envelope (legacy helper).
func Error(c *gin.Context, statusCode int, code string, message string, fields map[string]string) {
	c.Set("error_code", code)
	reqID := GetRequestID(c)
	c.JSON(statusCode, ErrorEnvelope{
		Error: ErrorBody{
			Code:    ErrorCode(code),
			Message: message,
			Fields:  fields,
		},
		RequestID: reqID,
	})
}

// InternalServerError is a helper for 500 responses without leaking details.
func InternalServerError(c *gin.Context) {
	Fail(c, CodeInternalError, "Terjadi kesalahan internal pada server")
}
