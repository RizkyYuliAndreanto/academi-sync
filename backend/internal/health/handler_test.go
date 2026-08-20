package health_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/response"
	"bimbingan-backend/internal/server"

	"github.com/gin-gonic/gin"
)

func TestLivenessAndReadinessHandlers(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			ServiceName: "test-service",
			Environment: "test",
		},
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := server.SetupRouter(cfg, logger, nil)

	t.Run("GET /health/live returns 200 OK without DB dependency", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/live", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got: %d", resp.Code)
		}

		var payload struct {
			Data struct {
				Status string `json:"status"`
			} `json:"data"`
			RequestID string `json:"request_id"`
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to unmarshal JSON response: %v", err)
		}

		if payload.Data.Status != "ok" {
			t.Fatalf("expected data.status to be 'ok', got: '%s'", payload.Data.Status)
		}

		if payload.RequestID == "" {
			t.Fatal("expected request_id to be populated in response payload")
		}
	})

	t.Run("GET /health/ready returns 200 OK when DB is disabled/nil", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health/ready", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK for ready check when db disabled, got: %d", resp.Code)
		}

		var payload struct {
			Data struct {
				Status string `json:"status"`
			} `json:"data"`
			RequestID string `json:"request_id"`
		}

		if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
			t.Fatalf("failed to unmarshal JSON response: %v", err)
		}

		if payload.Data.Status != "ready" {
			t.Fatalf("expected data.status to be 'ready', got: '%s'", payload.Data.Status)
		}
	})

	t.Run("GET /health/ready error envelope adheres to REST contract", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(response.RequestIDKey, "test-req-ready-fail")

		response.Fail(c, response.CodeServiceUnavailable, "Layanan belum siap")

		if w.Code != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503, got %d", w.Code)
		}

		var env response.ErrorEnvelope
		if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
			t.Fatalf("failed to unmarshal error envelope: %v", err)
		}
		if env.Error.Code != response.CodeServiceUnavailable {
			t.Fatalf("expected code SERVICE_UNAVAILABLE, got: %s", env.Error.Code)
		}
		if env.RequestID != "test-req-ready-fail" {
			t.Fatalf("expected request_id 'test-req-ready-fail', got: %s", env.RequestID)
		}
	})
}
