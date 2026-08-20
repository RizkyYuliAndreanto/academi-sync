package database

import (
	"errors"
	"fmt"
	"strings"
)

// ErrDatabaseDisabled is returned when trying to perform DB operations while database is disabled.
var ErrDatabaseDisabled = errors.New("database is disabled")

// SanitizeError strips sensitive DSN credentials or database connection strings from error messages.
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	if strings.Contains(errStr, "postgres://") || strings.Contains(errStr, "postgresql://") || strings.Contains(errStr, "user=") || strings.Contains(errStr, "password=") {
		return errors.New("database connection failed")
	}
	return fmt.Errorf("database operation failed: %w", err)
}
