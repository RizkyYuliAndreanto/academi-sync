package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bimbingan-backend/internal/app"
	"bimbingan-backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("configuration error: %v", err)
		os.Exit(1)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Printf("application initialization error: %v", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil {
		log.Printf("application runtime error: %v", err)
		os.Exit(1)
	}
}
