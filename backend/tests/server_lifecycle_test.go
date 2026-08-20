package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"bimbingan-backend/internal/app"
	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
)

func TestServerLifecycleAndRecovery(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			ServiceName: "bimbingan-backend-test",
			Environment: "test",
		},
		HTTP: config.HTTPConfig{
			Host:              "127.0.0.1",
			Port:              0,
			ReadHeaderTimeout: 2 * time.Second,
			ReadTimeout:       5 * time.Second,
			WriteTimeout:      5 * time.Second,
			IdleTimeout:       5 * time.Second,
			ShutdownTimeout:   3 * time.Second,
		},
	}

	application, err := app.New(cfg)
	if err != nil {
		t.Fatalf("failed to create application: %v", err)
	}

	// Add panic endpoint to test recovery middleware
	application.Router().GET("/test-panic", func(c *gin.Context) {
		panic("simulated critical internal panic")
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind listener to random port: %v", err)
	}
	application.SetListener(listener)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- application.Run(ctx)
	}()

	serverAddr := application.Addr()

	// 1. Test /health/live endpoint
	liveURL := fmt.Sprintf("http://%s/health/live", serverAddr)
	req, _ := http.NewRequest("GET", liveURL, nil)
	req.Header.Set("Cookie", "session_secret_cookie=abc123secret")
	req.Header.Set("Authorization", "Bearer super-secret-jwt-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("HTTP GET /health/live failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var livePayload struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(body, &livePayload); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if livePayload.Data.Status != "ok" || livePayload.RequestID == "" {
		t.Fatalf("unexpected live response: %s", string(body))
	}

	headerReqID := resp.Header.Get("X-Request-ID")
	if headerReqID == "" || headerReqID != livePayload.RequestID {
		t.Fatalf("mismatch between X-Request-ID header (%s) and JSON request_id (%s)", headerReqID, livePayload.RequestID)
	}

	// 2. Test Panic Recovery Middleware
	panicURL := fmt.Sprintf("http://%s/test-panic", serverAddr)
	panicResp, err := http.Get(panicURL)
	if err != nil {
		t.Fatalf("HTTP GET /test-panic failed: %v", err)
	}
	defer panicResp.Body.Close()

	if panicResp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500 on panic, got: %d", panicResp.StatusCode)
	}

	contentType := panicResp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected Content-Type application/json on panic response, got: %s", contentType)
	}

	panicHeaderID := panicResp.Header.Get("X-Request-ID")

	panicBody, _ := io.ReadAll(panicResp.Body)
	panicStr := string(panicBody)

	if strings.Contains(panicStr, "goroutine") || strings.Contains(panicStr, "simulated critical internal panic") {
		t.Fatalf("panic response body leaked stack trace or internal error string: %s", panicStr)
	}

	var panicPayload response.ErrorEnvelope
	if err := json.Unmarshal(panicBody, &panicPayload); err != nil {
		t.Fatalf("failed to parse panic error envelope: %v", err)
	}
	if panicPayload.Error.Code != response.CodeInternalError {
		t.Fatalf("expected error code INTERNAL_ERROR, got: %s", panicPayload.Error.Code)
	}
	if panicPayload.RequestID == "" || panicPayload.RequestID != panicHeaderID {
		t.Fatalf("expected request_id in panic response envelope (%s) to match header (%s)", panicPayload.RequestID, panicHeaderID)
	}

	// 3. Test Graceful Shutdown
	cancel() // Trigger context cancellation

	select {
	case runErr := <-errChan:
		if runErr != nil {
			t.Fatalf("application.Run returned error on graceful shutdown: %v", runErr)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("server shutdown timed out")
	}
}
