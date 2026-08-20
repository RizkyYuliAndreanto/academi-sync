package response_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/response"
	"bimbingan-backend/internal/server"

	"github.com/gin-gonic/gin"
)

func TestRESTContractExclusivityAndHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success envelope contains data and request_id but NOT error", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "req-exclusive-1")

		response.OK(c, map[string]string{"result": "ok"})

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		_, hasData := body["data"]
		_, hasError := body["error"]

		if hasData == hasError {
			t.Fatal("response envelope MUST contain exactly one of 'data' or 'error'")
		}
		if !hasData || hasError {
			t.Fatalf("expected data without error, got body: %v", body)
		}
		if body["request_id"] != "req-exclusive-1" {
			t.Fatalf("expected request_id 'req-exclusive-1', got: %v", body["request_id"])
		}
	})

	t.Run("error envelope contains error and request_id but NOT data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "req-exclusive-2")

		response.Fail(c, response.CodeForbidden, "Akses ditolak")

		var body map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("failed to parse JSON: %v", err)
		}

		_, hasData := body["data"]
		_, hasError := body["error"]

		if hasData == hasError {
			t.Fatal("response envelope MUST contain exactly one of 'data' or 'error'")
		}
		if !hasError || hasData {
			t.Fatalf("expected error without data, got body: %v", body)
		}
	})

	t.Run("NoContent 204 has no body envelope but carries status code 204", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		response.NoContent(c)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected status 204, got %d", w.Code)
		}
		if w.Body.Len() > 0 {
			t.Fatalf("expected 0 byte body for 204, got %d bytes: %s", w.Body.Len(), w.Body.String())
		}
	})
}

func TestNoRouteAndNoMethodIntegration(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			ServiceName: "contract-test-service",
			Environment: "test",
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := server.SetupRouter(cfg, logger, nil)

	t.Run("GET /non-existent-route returns 404 RESOURCE_NOT_FOUND JSON", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/non-existent-route", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusNotFound {
			t.Fatalf("expected status 404 for NoRoute, got: %d", resp.Code)
		}

		contentType := resp.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Fatalf("expected Content-Type application/json, got: %s", contentType)
		}

		var env response.ErrorEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to unmarshal 404 response: %v", err)
		}
		if env.Error.Code != response.CodeResourceNotFound {
			t.Fatalf("expected code RESOURCE_NOT_FOUND, got: %s", env.Error.Code)
		}
		if env.RequestID == "" {
			t.Fatal("expected request_id in 404 response")
		}

		headerReqID := resp.Header().Get("X-Request-ID")
		if headerReqID != env.RequestID {
			t.Fatalf("mismatch between header request_id (%s) and JSON request_id (%s)", headerReqID, env.RequestID)
		}
	})

	t.Run("POST /health/live returns 405 METHOD_NOT_ALLOWED JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/health/live", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected status 405 for NoMethod, got: %d", resp.Code)
		}

		var env response.ErrorEnvelope
		if err := json.Unmarshal(resp.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to unmarshal 405 response: %v", err)
		}
		if env.Error.Code != response.CodeMethodNotAllowed {
			t.Fatalf("expected code METHOD_NOT_ALLOWED, got: %s", env.Error.Code)
		}
	})
}

func TestSentinelNonLeakage(t *testing.T) {
	sentinels := []string{
		"database-password-sentinel-12345",
		"internal-stack-sentinel-goroutine-67890",
		"private-hostname-sentinel-internal.cluster.local",
	}

	gin.SetMode(gin.TestMode)

	for _, sentinel := range sentinels {
		t.Run("redacts sentinel "+sentinel, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Simulate internal server error with dangerous cause
			internalErr := response.NewAppError(
				response.CodeInternalError,
				"Terjadi kesalahan internal pada server",
				&customCauseError{detail: sentinel},
			)

			response.AppErr(c, internalErr)

			body, _ := io.ReadAll(w.Body)
			bodyStr := string(body)

			if strings.Contains(bodyStr, sentinel) {
				t.Fatalf("CRITICAL SECURITY VIOLATION: sentinel %q leaked in response body: %s", sentinel, bodyStr)
			}
		})
	}
}

type customCauseError struct {
	detail string
}

func (e *customCauseError) Error() string {
	return e.detail
}
