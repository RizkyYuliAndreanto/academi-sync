package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// List of sensitive environment variable keys for redaction.
var SensitiveKeys = map[string]bool{
	"DATABASE_URL":         true,
	"JWT_SECRET":           true,
	"REFRESH_TOKEN_SECRET": true,
	"MINIO_ACCESS_KEY":     true,
	"MINIO_SECRET_KEY":     true,
	"TURN_SHARED_SECRET":   true,
}

// InsecurePlaceholderSecrets lists dangerous placeholder secrets forbidden in production.
var InsecurePlaceholderSecrets = []string{
	"changeme",
	"secret",
	"12345",
	"admin",
	"password",
	"your-secret-key",
	"placeholder",
}

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Security SecurityConfig
	Database DatabaseConfig
	Storage  StorageConfig
	TURN     TURNConfig
}

type AppConfig struct {
	Environment string // "development", "test", "production"
	ServiceName string
	LogLevel    string // "debug", "info", "warn", "error"
	LogFormat   string // "text", "json"
}

type HTTPConfig struct {
	Host              string
	Port              int
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ShutdownTimeout   time.Duration
	AllowedOrigins    []string
}

type SecurityConfig struct {
	CookieSecure   bool
	CookieSameSite string // "lax", "strict", "none"
	CookieDomain   string
}

type DatabaseConfig struct {
	Enabled           bool
	URL               string
	MaxConnections    int32
	MinConnections    int32
	ConnectTimeout    time.Duration
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

type StorageConfig struct {
	Enabled   bool
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type TURNConfig struct {
	Enabled       bool
	Host          string
	SharedSecret  string
	CredentialTTL time.Duration
}

// LookupEnv represents an abstraction for looking up environment variables.
type LookupEnv func(key string) (string, bool)

// Load reads configuration using standard os.LookupEnv.
func Load() (*Config, error) {
	return LoadFrom(nil)
}

// LoadFrom reads configuration using the provided LookupEnv function.
func LoadFrom(lookup LookupEnv) (*Config, error) {
	if lookup == nil {
		lookup = os.LookupEnv
	}

	getEnv := func(key, fallback string) string {
		if val, ok := lookup(key); ok && strings.TrimSpace(val) != "" {
			return strings.TrimSpace(val)
		}
		return fallback
	}

	getEnvBool := func(key string, fallback bool) (bool, error) {
		valStr := getEnv(key, "")
		if valStr == "" {
			return fallback, nil
		}
		b, err := strconv.ParseBool(valStr)
		if err != nil {
			return false, fmt.Errorf("configuration error: %s must be a valid boolean", key)
		}
		return b, nil
	}

	getEnvInt := func(key string, fallback int) (int, error) {
		valStr := getEnv(key, "")
		if valStr == "" {
			return fallback, nil
		}
		val, err := strconv.Atoi(valStr)
		if err != nil {
			return 0, fmt.Errorf("configuration error: %s must be a valid integer", key)
		}
		return val, nil
	}

	getEnvInt32 := func(key string, fallback int32) (int32, error) {
		val, err := getEnvInt(key, int(fallback))
		if err != nil {
			return 0, err
		}
		return int32(val), nil
	}

	getEnvDuration := func(key string, fallback time.Duration) (time.Duration, error) {
		valStr := getEnv(key, "")
		if valStr == "" {
			return fallback, nil
		}
		val, err := time.ParseDuration(valStr)
		if err != nil {
			return 0, fmt.Errorf("configuration error: %s must be a valid duration (e.g., 5s, 10m)", key)
		}
		return val, nil
	}

	// App
	env := getEnv("APP_ENV", "development")
	serviceName := getEnv("APP_SERVICE_NAME", "bimbingan-backend")
	logLevel := getEnv("LOG_LEVEL", "debug")
	logFormat := getEnv("LOG_FORMAT", "text")

	// HTTP
	host := getEnv("HTTP_HOST", "0.0.0.0")
	port, err := getEnvInt("HTTP_PORT", 8080)
	if err != nil {
		return nil, err
	}
	readHeaderTimeout, err := getEnvDuration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	readTimeout, err := getEnvDuration("HTTP_READ_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, err
	}
	writeTimeout, err := getEnvDuration("HTTP_WRITE_TIMEOUT", 30*time.Second)
	if err != nil {
		return nil, err
	}
	idleTimeout, err := getEnvDuration("HTTP_IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, err
	}
	shutdownTimeout, err := getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, err
	}

	originsRaw := getEnv("HTTP_ALLOWED_ORIGINS", "http://localhost:5173")
	var allowedOrigins []string
	if originsRaw != "" {
		for _, o := range strings.Split(originsRaw, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	// Security
	cookieSecure, err := getEnvBool("COOKIE_SECURE", false)
	if err != nil {
		return nil, err
	}
	cookieSameSite := getEnv("COOKIE_SAME_SITE", "lax")
	cookieDomain := getEnv("COOKIE_DOMAIN", "")

	// Database
	dbEnabled, err := getEnvBool("DATABASE_ENABLED", false)
	if err != nil {
		return nil, err
	}
	dbURL := getEnv("DATABASE_URL", "")
	dbMaxConn, err := getEnvInt32("DATABASE_MAX_CONNECTIONS", 10)
	if err != nil {
		return nil, err
	}
	dbMinConn, err := getEnvInt32("DATABASE_MIN_CONNECTIONS", 1)
	if err != nil {
		return nil, err
	}
	dbConnTimeout, err := getEnvDuration("DATABASE_CONNECT_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}
	dbMaxConnLifetime, err := getEnvDuration("DATABASE_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return nil, err
	}
	dbMaxConnIdleTime, err := getEnvDuration("DATABASE_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return nil, err
	}
	dbHealthCheckPeriod, err := getEnvDuration("DATABASE_HEALTH_CHECK_PERIOD", 30*time.Second)
	if err != nil {
		return nil, err
	}

	// Storage (MinIO)
	storageEnabled, err := getEnvBool("STORAGE_ENABLED", false)
	if err != nil {
		return nil, err
	}
	minioEndpoint := getEnv("MINIO_ENDPOINT", "")
	minioAccessKey := getEnv("MINIO_ACCESS_KEY", "")
	minioSecretKey := getEnv("MINIO_SECRET_KEY", "")
	minioBucket := getEnv("MINIO_BUCKET", "")
	minioUseSSL, err := getEnvBool("MINIO_USE_SSL", false)
	if err != nil {
		return nil, err
	}

	// TURN
	turnEnabled, err := getEnvBool("TURN_ENABLED", false)
	if err != nil {
		return nil, err
	}
	turnHost := getEnv("TURN_HOST", "")
	turnSecret := getEnv("TURN_SHARED_SECRET", "")
	turnTTL, err := getEnvDuration("TURN_CREDENTIAL_TTL", 10*time.Minute)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		App: AppConfig{
			Environment: env,
			ServiceName: serviceName,
			LogLevel:    logLevel,
			LogFormat:   logFormat,
		},
		HTTP: HTTPConfig{
			Host:              host,
			Port:              port,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
			ShutdownTimeout:   shutdownTimeout,
			AllowedOrigins:    allowedOrigins,
		},
		Security: SecurityConfig{
			CookieSecure:   cookieSecure,
			CookieSameSite: cookieSameSite,
			CookieDomain:   cookieDomain,
		},
		Database: DatabaseConfig{
			Enabled:           dbEnabled,
			URL:               dbURL,
			MaxConnections:    dbMaxConn,
			MinConnections:    dbMinConn,
			ConnectTimeout:    dbConnTimeout,
			MaxConnLifetime:   dbMaxConnLifetime,
			MaxConnIdleTime:   dbMaxConnIdleTime,
			HealthCheckPeriod: dbHealthCheckPeriod,
		},
		Storage: StorageConfig{
			Enabled:   storageEnabled,
			Endpoint:  minioEndpoint,
			AccessKey: minioAccessKey,
			SecretKey: minioSecretKey,
			Bucket:    minioBucket,
			UseSSL:    minioUseSSL,
		},
		TURN: TURNConfig{
			Enabled:       turnEnabled,
			Host:          turnHost,
			SharedSecret:  turnSecret,
			CredentialTTL: turnTTL,
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	// 1. Strict Environment Check
	switch c.App.Environment {
	case "development", "test", "production":
		// valid
	default:
		return fmt.Errorf("configuration error: APP_ENV must be one of [development, test, production], got '%s'", c.App.Environment)
	}

	// 2. Service Name Check
	if strings.TrimSpace(c.App.ServiceName) == "" {
		return errors.New("configuration error: APP_SERVICE_NAME cannot be empty")
	}

	// 3. HTTP Port Check
	if c.HTTP.Port < 1 || c.HTTP.Port > 65535 {
		return fmt.Errorf("configuration error: HTTP_PORT must be between 1 and 65535, got %d", c.HTTP.Port)
	}

	// 4. HTTP Timeouts Check
	if c.HTTP.ReadHeaderTimeout <= 0 || c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.IdleTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 {
		return errors.New("configuration error: all HTTP timeouts must be positive durations (> 0)")
	}

	// 5. Database Enabled Validation
	if c.Database.Enabled {
		if strings.TrimSpace(c.Database.URL) == "" {
			return errors.New("configuration validation failed: DATABASE_URL is required when DATABASE_ENABLED is true")
		}
		if isPlaceholderSecret(c.Database.URL) {
			return errors.New("configuration validation failed: DATABASE_URL contains default placeholder value")
		}
		if c.Database.MaxConnections <= 0 {
			return errors.New("configuration validation failed: DATABASE_MAX_CONNECTIONS must be positive (> 0)")
		}
		if c.Database.MinConnections < 0 {
			return errors.New("configuration validation failed: DATABASE_MIN_CONNECTIONS must be non-negative (>= 0)")
		}
		if c.Database.MinConnections > c.Database.MaxConnections {
			return errors.New("configuration validation failed: DATABASE_MIN_CONNECTIONS cannot exceed DATABASE_MAX_CONNECTIONS")
		}
		if c.Database.ConnectTimeout <= 0 || c.Database.MaxConnLifetime <= 0 || c.Database.MaxConnIdleTime <= 0 || c.Database.HealthCheckPeriod <= 0 {
			return errors.New("configuration validation failed: all database duration settings must be positive (> 0)")
		}
	}

	// 6. Storage Enabled Validation
	if c.Storage.Enabled {
		if strings.TrimSpace(c.Storage.Endpoint) == "" {
			return errors.New("configuration validation failed: MINIO_ENDPOINT is required when STORAGE_ENABLED is true")
		}
		if strings.TrimSpace(c.Storage.AccessKey) == "" {
			return errors.New("configuration validation failed: MINIO_ACCESS_KEY is required when STORAGE_ENABLED is true")
		}
		if strings.TrimSpace(c.Storage.SecretKey) == "" {
			return errors.New("configuration validation failed: MINIO_SECRET_KEY is required when STORAGE_ENABLED is true")
		}
		if isPlaceholderSecret(c.Storage.SecretKey) {
			return errors.New("configuration validation failed: MINIO_SECRET_KEY contains default placeholder value")
		}
		if strings.TrimSpace(c.Storage.Bucket) == "" {
			return errors.New("configuration validation failed: MINIO_BUCKET is required when STORAGE_ENABLED is true")
		}
	}

	// 7. TURN Enabled Validation
	if c.TURN.Enabled {
		if strings.TrimSpace(c.TURN.Host) == "" {
			return errors.New("configuration validation failed: TURN_HOST is required when TURN_ENABLED is true")
		}
		if strings.TrimSpace(c.TURN.SharedSecret) == "" {
			return errors.New("configuration validation failed: TURN_SHARED_SECRET is required when TURN_ENABLED is true")
		}
		if isPlaceholderSecret(c.TURN.SharedSecret) {
			return errors.New("configuration validation failed: TURN_SHARED_SECRET contains default placeholder value")
		}
	}

	// 8. Fail-Closed Production Rules
	if c.App.Environment == "production" {
		if c.App.LogFormat != "json" {
			return fmt.Errorf("configuration validation failed: LOG_FORMAT must be 'json' when APP_ENV is production, got '%s'", c.App.LogFormat)
		}
		if !c.Security.CookieSecure {
			return errors.New("configuration validation failed: COOKIE_SECURE must be true when APP_ENV is production")
		}
		if !c.Database.Enabled {
			return errors.New("configuration validation failed: DATABASE_ENABLED must be true when APP_ENV is production")
		}
		for _, origin := range c.HTTP.AllowedOrigins {
			if origin == "*" {
				return errors.New("configuration validation failed: HTTP_ALLOWED_ORIGINS cannot contain wildcard '*' when APP_ENV is production")
			}
		}
	}

	return nil
}

// SafeSummary returns a safe map for logging startup info without leaking secrets or URLs.
func (c *Config) SafeSummary() map[string]any {
	return map[string]any{
		"environment":     c.App.Environment,
		"service":         c.App.ServiceName,
		"http_port":       c.HTTP.Port,
		"cookie_secure":   c.Security.CookieSecure,
		"database_enable": c.Database.Enabled,
		"storage_enable":  c.Storage.Enabled,
		"turn_enable":     c.TURN.Enabled,
	}
}

func isPlaceholderSecret(val string) bool {
	lower := strings.ToLower(val)
	for _, placeholder := range InsecurePlaceholderSecrets {
		if lower == placeholder || strings.Contains(lower, placeholder) {
			return true
		}
	}
	return false
}
