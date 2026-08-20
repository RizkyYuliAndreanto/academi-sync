package main

import (
	"fmt"
	"log"
	"os"

	"bimbingan-backend/internal/config"
	"bimbingan-backend/internal/migration"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  migrate up        - Run all pending up migrations")
		fmt.Println("  migrate version   - Print current migration version and dirty status")
		fmt.Println("  migrate down <n>  - Rollback <n> migration steps")
		os.Exit(1)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	if cfg.Database.URL == "" {
		log.Fatalf("DATABASE_URL environment variable is required to run migrations")
	}

	runner := migration.NewRunner(cfg.Database.URL)
	cmd := os.Args[1]

	switch cmd {
	case "up":
		log.Println("running database migrations up...")
		if err := runner.Up(); err != nil {
			log.Fatalf("migration up failed: %v", err)
		}
		version, dirty, _ := runner.Version()
		log.Printf("migration up completed successfully. current version: %d, dirty: %v\n", version, dirty)

	case "version":
		version, dirty, err := runner.Version()
		if err != nil {
			log.Fatalf("failed to get migration version: %v", err)
		}
		log.Printf("current migration version: %d, dirty: %v\n", version, dirty)

	case "down":
		if len(os.Args) < 3 {
			log.Fatalf("usage: migrate down <steps> (e.g. migrate down 1)")
		}
		steps, err := migration.ParseSteps(os.Args[2])
		if err != nil {
			log.Fatalf("invalid steps argument: %v", err)
		}

		log.Printf("rolling back %d migration step(s)...\n", steps)
		if err := runner.Down(steps); err != nil {
			log.Fatalf("migration down failed: %v", err)
		}
		version, dirty, _ := runner.Version()
		log.Printf("migration down completed. current version: %d, dirty: %v\n", version, dirty)

	default:
		log.Fatalf("unknown command '%s'. Supported commands: up, version, down", cmd)
	}
}
