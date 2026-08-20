package health

import (
	"time"

	"bimbingan-backend/internal/database"
	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

// NewHandler initializes a health Handler with an optional database connection pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{
		pool: pool,
	}
}

// Live handles Liveness probe requests (GET /health/live).
// It verifies that the application process is running. It does NOT depend on PostgreSQL.
func (h *Handler) Live(c *gin.Context) {
	response.OK(c, gin.H{
		"status": "ok",
	})
}

// Ready handles Readiness probe requests (GET /health/ready).
// It verifies that core dependencies (PostgreSQL) are accessible.
func (h *Handler) Ready(c *gin.Context) {
	if err := database.CheckHealth(c.Request.Context(), h.pool, 2*time.Second); err != nil {
		response.Fail(c, response.CodeServiceUnavailable, "Layanan belum siap")
		return
	}

	response.OK(c, gin.H{
		"status": "ready",
	})
}
