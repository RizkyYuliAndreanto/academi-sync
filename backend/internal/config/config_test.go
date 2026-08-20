package config_test

import (
	"fmt"
	"strings"
	"testing"

	"bimbingan-backend/internal/config"
)

func mapLookup(env map[string]string) config.LookupEnv {
	return func(key string) (string, bool) {
		val, ok := env[key]
		return val, ok
	}
}

func TestConfigParsingAndValidation(t *testing.T) {
	t.Run("valid development defaults", func(t *testing.T) {
		cfg, err := config.LoadFrom(mapLookup(map[string]string{
			"APP_ENV": "development",
		}))
		if err != nil {
			t.Fatalf("expected successful config load, got error: %v", err)
		}
		if cfg.App.Environment != "development" {
			t.Fatalf("expected APP_ENV to be development, got: %s", cfg.App.Environment)
		}
		if cfg.HTTP.Port != 8080 {
			t.Fatalf("expected default HTTP_PORT to be 8080, got: %d", cfg.HTTP.Port)
		}
	})

	t.Run("fails on unknown environment typo (e.g. prodution)", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"APP_ENV": "prodution",
		}))
		if err == nil {
			t.Fatal("expected error for typo in APP_ENV, got nil")
		}
		if !strings.Contains(err.Error(), "APP_ENV must be one of") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("fails on non-integer port", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"HTTP_PORT": "invalid-port",
		}))
		if err == nil {
			t.Fatal("expected error for non-integer HTTP_PORT, got nil")
		}
	})

	t.Run("fails on out-of-range port", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"HTTP_PORT": "70000",
		}))
		if err == nil {
			t.Fatal("expected error for out-of-range HTTP_PORT, got nil")
		}
	})

	t.Run("fails on invalid duration format", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"HTTP_READ_TIMEOUT": "10xyz",
		}))
		if err == nil {
			t.Fatal("expected error for invalid duration, got nil")
		}
	})

	t.Run("fails on non-positive HTTP timeout", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"HTTP_READ_TIMEOUT": "0s",
		}))
		if err == nil {
			t.Fatal("expected error for 0s HTTP_READ_TIMEOUT, got nil")
		}
	})

	t.Run("parses multiple allowed origins cleanly", func(t *testing.T) {
		cfg, err := config.LoadFrom(mapLookup(map[string]string{
			"HTTP_ALLOWED_ORIGINS": "http://localhost:5173, https://app.example.com ",
		}))
		if err != nil {
			t.Fatalf("unexpected error parsing origins: %v", err)
		}
		if len(cfg.HTTP.AllowedOrigins) != 2 {
			t.Fatalf("expected 2 origins, got %d", len(cfg.HTTP.AllowedOrigins))
		}
		if cfg.HTTP.AllowedOrigins[1] != "https://app.example.com" {
			t.Fatalf("expected trimmed origin 'https://app.example.com', got '%s'", cfg.HTTP.AllowedOrigins[1])
		}
	})

	t.Run("fails on invalid boolean string", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(map[string]string{
			"COOKIE_SECURE": "not-a-bool",
		}))
		if err == nil {
			t.Fatal("expected error for invalid boolean, got nil")
		}
	})
}

func TestProductionFailClosedPolicy(t *testing.T) {
	validProdEnv := map[string]string{
		"APP_ENV":              "production",
		"APP_SERVICE_NAME":     "bimbingan-backend",
		"LOG_FORMAT":           "json",
		"COOKIE_SECURE":        "true",
		"HTTP_ALLOWED_ORIGINS": "https://bimbingan.example.com",
		"DATABASE_ENABLED":     "true",
		"DATABASE_URL":         "postgres://prod_user:prod_pass@localhost:5432/bimbingan_prod",
	}

	t.Run("accepts valid production config", func(t *testing.T) {
		_, err := config.LoadFrom(mapLookup(validProdEnv))
		if err != nil {
			t.Fatalf("expected valid production config to pass, got: %v", err)
		}
	})

	t.Run("rejects production with DATABASE_ENABLED=false", func(t *testing.T) {
		env := map[string]string{}
		for k, v := range validProdEnv {
			env[k] = v
		}
		env["DATABASE_ENABLED"] = "false"

		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for DATABASE_ENABLED=false in production")
		}
		if !strings.Contains(err.Error(), "DATABASE_ENABLED must be true") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects production with COOKIE_SECURE=false", func(t *testing.T) {
		env := map[string]string{}
		for k, v := range validProdEnv {
			env[k] = v
		}
		env["COOKIE_SECURE"] = "false"

		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for COOKIE_SECURE=false in production")
		}
		if !strings.Contains(err.Error(), "COOKIE_SECURE must be true") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects production with LOG_FORMAT=text", func(t *testing.T) {
		env := map[string]string{}
		for k, v := range validProdEnv {
			env[k] = v
		}
		env["LOG_FORMAT"] = "text"

		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for LOG_FORMAT=text in production")
		}
	})

	t.Run("rejects production with wildcard origin '*'", func(t *testing.T) {
		env := map[string]string{}
		for k, v := range validProdEnv {
			env[k] = v
		}
		env["HTTP_ALLOWED_ORIGINS"] = "*"

		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for wildcard origin in production")
		}
		if !strings.Contains(err.Error(), "cannot contain wildcard") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects DATABASE_ENABLED=true without DATABASE_URL", func(t *testing.T) {
		env := map[string]string{
			"DATABASE_ENABLED": "true",
			"DATABASE_URL":     "",
		}
		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for enabled database without URL")
		}
	})

	t.Run("rejects STORAGE_ENABLED=true with placeholder secret", func(t *testing.T) {
		env := map[string]string{
			"STORAGE_ENABLED":  "true",
			"MINIO_ENDPOINT":   "minio:9000",
			"MINIO_ACCESS_KEY": "admin",
			"MINIO_SECRET_KEY": "changeme",
			"MINIO_BUCKET":     "pdf-docs",
		}
		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for storage with placeholder secret")
		}
		if !strings.Contains(err.Error(), "placeholder value") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("rejects TURN_ENABLED=true without shared secret", func(t *testing.T) {
		env := map[string]string{
			"TURN_ENABLED": "true",
			"TURN_HOST":    "turn.example.com",
		}
		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for TURN without secret")
		}
	})
}

func TestSecretRedaction(t *testing.T) {
	sentinel := "my-super-confidential-sentinel-key-9999"

	t.Run("never leaks secret value in validation error message", func(t *testing.T) {
		env := map[string]string{
			"STORAGE_ENABLED":  "true",
			"MINIO_ENDPOINT":   "minio:9000",
			"MINIO_ACCESS_KEY": "access",
			"MINIO_SECRET_KEY": sentinel,
			"MINIO_BUCKET":     "", // missing bucket triggers error
		}

		_, err := config.LoadFrom(mapLookup(env))
		if err == nil {
			t.Fatal("expected error for missing bucket")
		}

		errMsg := err.Error()
		if strings.Contains(errMsg, sentinel) {
			t.Fatalf("CRITICAL SECURITY RISK: secret sentinel value leaked in error message: %s", errMsg)
		}
	})

	t.Run("SafeSummary does not leak sensitive credentials", func(t *testing.T) {
		cfg, err := config.LoadFrom(mapLookup(map[string]string{
			"DATABASE_ENABLED": "true",
			"DATABASE_URL":     "postgres://user:" + sentinel + "@localhost:5432/db",
		}))
		if err != nil {
			t.Fatalf("unexpected load error: %v", err)
		}

		summary := cfg.SafeSummary()
		summaryStr := fmt.Sprintf("%v", summary)

		if strings.Contains(summaryStr, sentinel) {
			t.Fatalf("CRITICAL SECURITY RISK: secret sentinel value leaked in SafeSummary: %s", summaryStr)
		}
	})
}
