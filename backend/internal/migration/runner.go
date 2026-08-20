package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"bimbingan-backend/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
)

type Runner struct {
	dbURL string
}

func NewRunner(dbURL string) *Runner {
	return &Runner{dbURL: dbURL}
}

func (r *Runner) newMigrate() (*migrate.Migrate, *sql.DB, error) {
	if r.dbURL == "" {
		return nil, nil, errors.New("database URL is required for migration")
	}

	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create migration source driver: %w", err)
	}

	db, err := sql.Open("postgres", r.dbURL)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sql database for migration: %w", err)
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create postgres migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to initialize migrate instance: %w", err)
	}

	return m, db, nil
}

func (r *Runner) Up() error {
	m, db, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration up failed: %w", err)
	}
	return nil
}

func (r *Runner) Down(steps int) error {
	if steps <= 0 {
		return errors.New("steps for migration down must be greater than 0")
	}

	m, db, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := m.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration down failed: %w", err)
	}
	return nil
}

func (r *Runner) DownAll() error {
	m, db, err := r.newMigrate()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration down all failed: %w", err)
	}
	return nil
}

func (r *Runner) Version() (uint, bool, error) {
	m, db, err := r.newMigrate()
	if err != nil {
		return 0, false, err
	}
	defer db.Close()

	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, fmt.Errorf("failed to get migration version: %w", err)
	}
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, nil
}

// ParseSteps parses string argument to integer for down command.
func ParseSteps(arg string) (int, error) {
	steps, err := strconv.Atoi(arg)
	if err != nil || steps <= 0 {
		return 0, fmt.Errorf("invalid step argument '%s': must be a positive integer", arg)
	}
	return steps, nil
}
