package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/database"
)

func TestPoolDisabledMode(t *testing.T) {
	cfg := config.DatabaseConfig{
		Enabled: false,
	}

	pool, err := database.NewPool(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("expected no error when database is disabled, got: %v", err)
	}
	if pool != nil {
		t.Fatal("expected nil pool when database is disabled")
	}

	// Readiness check should return nil when pool is nil
	if err := database.CheckHealth(context.Background(), pool, 1*time.Second); err != nil {
		t.Fatalf("expected healthy check when database is disabled, got: %v", err)
	}
}

func TestPoolInvalidDSNRedaction(t *testing.T) {
	sensitiveURL := "postgres://secret_user:super_secret_password_12345@invalid-host-name-xyz:5432/db"
	cfg := config.DatabaseConfig{
		Enabled:           true,
		URL:               sensitiveURL,
		MaxConnections:    5,
		MinConnections:    1,
		ConnectTimeout:    50 * time.Millisecond,
		MaxConnLifetime:   1 * time.Minute,
		MaxConnIdleTime:   1 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}

	_, err := database.NewPool(context.Background(), cfg, nil)
	if err == nil {
		t.Fatal("expected error connecting to invalid database host, got nil")
	}

	errStr := err.Error()
	if strings.Contains(errStr, "super_secret_password_12345") || strings.Contains(errStr, sensitiveURL) {
		t.Fatalf("CRITICAL SECURITY RISK: DSN credentials leaked in error message: %s", errStr)
	}
}

func TestSanitizeError(t *testing.T) {
	t.Run("nil error returns nil", func(t *testing.T) {
		if err := database.SanitizeError(nil); err != nil {
			t.Fatalf("expected nil, got %v", err)
		}
	})

	t.Run("sanitizes postgres connection URL", func(t *testing.T) {
		err := errors.New("failed to connect to postgres://admin:secretPass@localhost:5432/mydb")
		sanitized := database.SanitizeError(err)
		if strings.Contains(sanitized.Error(), "secretPass") {
			t.Fatalf("failed to redact password: %v", sanitized)
		}
	})
}
