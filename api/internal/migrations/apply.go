package migrations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/postgres"
)

const (
	applyAllMigrations = 0
	lockName           = "auth-api-migrations"
	lockTimeout        = 30 * time.Second
)

// Apply runs all pending Atlas migrations before the API starts serving traffic.
func Apply(ctx context.Context, db *sql.DB) error {
	dir, err := EmbeddedDir()
	if err != nil {
		return err
	}
	if err := migrate.Validate(dir); err != nil {
		return fmt.Errorf("validate migration directory: %w", err)
	}

	drv, err := postgres.Open(db)
	if err != nil {
		return fmt.Errorf("open atlas postgres driver: %w", err)
	}
	unlock, err := drv.Lock(ctx, lockName, lockTimeout)
	if err != nil {
		return fmt.Errorf("lock migrations: %w", err)
	}
	defer unlock()

	revisionStore, err := NewRevisionStore(ctx, db)
	if err != nil {
		return fmt.Errorf("initialize migration revisions: %w", err)
	}
	executor, err := migrate.NewExecutor(drv, dir, revisionStore)
	if err != nil {
		return fmt.Errorf("create migration executor: %w", err)
	}
	if err := executor.ExecuteN(ctx, applyAllMigrations); err != nil && !errors.Is(err, migrate.ErrNoPendingFiles) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
