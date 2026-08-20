package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/database"
	"bimbingan-backend/internal/server"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	db         *pgxpool.Pool
	httpServer *http.Server
	ginEngine  *gin.Engine
	listener   net.Listener
}

// New instantiates a new App with the given configuration and logger setup.
func New(cfg *config.Config) (*App, error) {
	var handler slog.Handler
	if cfg.App.Environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}
	logger := slog.New(handler)

	// Initialize database pool (returns nil, nil if Database.Enabled is false)
	db, err := database.NewPool(context.Background(), cfg.Database, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database pool: %w", err)
	}

	ginEngine := server.SetupRouter(cfg, logger, db)
	httpSrv := server.NewHTTPServer(cfg, ginEngine)

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         db,
		httpServer: httpSrv,
		ginEngine:  ginEngine,
	}, nil
}

// SetListener sets a custom listener for testing purposes (e.g. random port).
func (a *App) SetListener(l net.Listener) {
	a.listener = l
}

// Addr returns the bound address of the server.
func (a *App) Addr() string {
	if a.listener != nil {
		return a.listener.Addr().String()
	}
	return a.httpServer.Addr
}

// Logger returns the application logger instance.
func (a *App) Logger() *slog.Logger {
	return a.logger
}

// Router returns the Gin engine instance.
func (a *App) Router() *gin.Engine {
	return a.ginEngine
}

// DB returns the pgxpool connection pool instance.
func (a *App) DB() *pgxpool.Pool {
	return a.db
}

// Run starts the HTTP server and blocks until the context is canceled, performing graceful shutdown.
func (a *App) Run(ctx context.Context) error {
	serverErrChan := make(chan error, 1)

	go func() {
		a.logger.Info("starting HTTP server",
			slog.String("service", a.cfg.App.ServiceName),
			slog.String("environment", a.cfg.App.Environment),
			slog.String("addr", a.Addr()),
			slog.Any("config_summary", a.cfg.SafeSummary()),
		)

		var err error
		if a.listener != nil {
			err = a.httpServer.Serve(a.listener)
		} else {
			err = a.httpServer.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- fmt.Errorf("http server ListenAndServe error: %w", err)
		}
		close(serverErrChan)
	}()

	select {
	case err := <-serverErrChan:
		if a.db != nil {
			a.db.Close()
		}
		return err
	case <-ctx.Done():
		a.logger.Info("shutdown signal received, shutting down HTTP server gracefully...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		a.logger.Error("server forced to shutdown", slog.String("error", err.Error()))
		if a.db != nil {
			a.db.Close()
		}
		return fmt.Errorf("server shutdown forced: %w", err)
	}

	// Close database connection pool after HTTP server has finished active requests
	if a.db != nil {
		a.logger.Info("closing database connection pool...")
		a.db.Close()
	}

	a.logger.Info("server exited cleanly")
	return nil
}
