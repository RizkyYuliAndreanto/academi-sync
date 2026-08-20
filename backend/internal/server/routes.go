package server

import (
	"log/slog"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/health"
	"bimbingan-backend/internal/middleware"
	"bimbingan-backend/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupRouter initializes the Gin engine with standard middleware, routes, and optional db pool.
func SetupRouter(cfg *config.Config, logger *slog.Logger, db *pgxpool.Pool) *gin.Engine {
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else if cfg.App.Environment == "test" {
		gin.SetMode(gin.TestMode)
	}

	router := gin.New()
	router.HandleMethodNotAllowed = true

	// Trusted proxy configuration
	_ = router.SetTrustedProxies(nil)

	// Order of middleware:
	// 1. Request ID (generates/validates X-Request-ID and sets context)
	// 2. Recovery (catches panics and includes Request ID in panic error)
	// 3. Security Headers
	// 4. Access Log (logs request duration, status, and request_id)
	router.Use(middleware.RequestID())
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.AccessLog(logger, cfg.App.ServiceName, cfg.App.Environment))

	// NoRoute (404) and NoMethod (405) handlers adhering to JSON contract
	router.NoRoute(func(c *gin.Context) {
		response.Fail(c, response.CodeResourceNotFound, "Endpoint tidak ditemukan")
	})

	router.NoMethod(func(c *gin.Context) {
		response.Fail(c, response.CodeMethodNotAllowed, "Metode HTTP tidak diizinkan")
	})

	// Health routes
	healthHandler := health.NewHandler(db)
	healthGroup := router.Group("/health")
	{
		healthGroup.GET("/live", healthHandler.Live)
		healthGroup.GET("/ready", healthHandler.Ready)
	}

	return router
}
