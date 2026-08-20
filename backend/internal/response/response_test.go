package response_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
)

func TestStatusForCodeMapping(t *testing.T) {
	tests := []struct {
		code   response.ErrorCode
		status int
	}{
		{response.CodeValidationError, http.StatusBadRequest},
		{response.CodeInvalidRequest, http.StatusBadRequest},
		{response.CodeUnauthenticated, http.StatusUnauthorized},
		{response.CodeForbidden, http.StatusForbidden},
		{response.CodeResourceNotFound, http.StatusNotFound},
		{response.CodeMethodNotAllowed, http.StatusMethodNotAllowed},
		{response.CodeResourceConflict, http.StatusConflict},
		{response.CodePayloadTooLarge, http.StatusRequestEntityTooLarge},
		{response.CodeUnsupportedMediaType, http.StatusUnsupportedMediaType},
		{response.CodeRateLimited, http.StatusTooManyRequests},
		{response.CodeInternalError, http.StatusInternalServerError},
		{response.CodeServiceUnavailable, http.StatusServiceUnavailable},
		{response.ErrorCode("UNKNOWN_CUSTOM_CODE"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			got := response.StatusForCode(tt.code)
			if got != tt.status {
				t.Errorf("StatusForCode(%s) = %d; want %d", tt.code, got, tt.status)
			}
		})
	}
}

func TestResponseHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("OK helper", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "test-req-123")

		response.OK(c, map[string]string{"message": "success"})

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got: %d", w.Code)
		}

		var envelope response.SuccessEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if envelope.RequestID != "test-req-123" {
			t.Fatalf("expected request_id 'test-req-123', got: '%s'", envelope.RequestID)
		}
	})

	t.Run("Created helper", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "test-req-123")

		response.Created(c, map[string]string{"id": "doc-1"})

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201, got: %d", w.Code)
		}
	})

	t.Run("ValidationFail helper", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "test-req-123")

		fields := map[string]string{"email": "invalid email format"}
		response.ValidationFail(c, "Data tidak valid", fields)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got: %d", w.Code)
		}

		var envelope response.ErrorEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}
		if envelope.Error.Code != response.CodeValidationError {
			t.Fatalf("expected code VALIDATION_ERROR, got: %s", envelope.Error.Code)
		}
		if envelope.Error.Fields["email"] != "invalid email format" {
			t.Fatalf("expected field email error, got: %v", envelope.Error.Fields)
		}
	})

	t.Run("NoContent helper", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		response.NoContent(c)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got: %d", w.Code)
		}
		if w.Body.Len() != 0 {
			t.Fatalf("expected empty body for 204 No Content, got %d bytes", w.Body.Len())
		}
	})

	t.Run("AppErr helper redacts internal cause", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "test-req-123")

		appErr := response.NewAppError(
			response.CodeResourceNotFound,
			"Resource tidak ditemukan",
			errors.New("db query SELECT * FROM secret_table failed: connection refused"),
		)

		response.AppErr(c, appErr)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status 404, got: %d", w.Code)
		}

		bodyStr := w.Body.String()
		if strings.Contains(bodyStr, "secret_table") || strings.Contains(bodyStr, "connection refused") {
			t.Fatalf("CRITICAL SECURITY RISK: internal error cause leaked in response: %s", bodyStr)
		}
	})
}
