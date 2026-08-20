package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DefaultReadinessTimeout is the maximum duration allowed for a database readiness ping.
const DefaultReadinessTimeout = 2 * time.Second

// CheckHealth pings the database pool with a short timeout.
// If pool is nil (disabled), it returns nil (healthy for disabled state in dev/test).
func CheckHealth(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) error {
	if pool == nil {
		return nil
	}
	if timeout <= 0 {
		timeout = DefaultReadinessTimeout
	}

	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return SanitizeError(err)
	}
	return nil
}
