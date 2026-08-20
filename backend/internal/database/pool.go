package database

import (
	"context"
	"fmt"
	"log/slog"

	"bimbingan-backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool initializes a new PostgreSQL connection pool using pgxpool.
// If cfg.Enabled is false, it returns nil, nil.
func NewPool(ctx context.Context, cfg config.DatabaseConfig, logger *slog.Logger) (*pgxpool.Pool, error) {
	if !cfg.Enabled {
		if logger != nil {
			logger.Info("database pool is disabled by configuration")
		}
		return nil, nil
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("database configuration parse failed: %w", SanitizeError(err))
	}

	// Apply typed configuration settings to pgxpool
	poolCfg.MaxConns = cfg.MaxConnections
	poolCfg.MinConns = cfg.MinConnections
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	// Connect with timeout
	connCtx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(connCtx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("database pool creation failed: %w", SanitizeError(err))
	}

	// Ping database to verify connection
	if err := pool.Ping(connCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", SanitizeError(err))
	}

	if logger != nil {
		logger.Info("database connection pool initialized successfully",
			slog.Int("max_conns", int(cfg.MaxConnections)),
			slog.Int("min_conns", int(cfg.MinConnections)),
		)
	}

	return pool, nil
}
