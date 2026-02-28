package migrate

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type PostgresLocker struct {
	retryInterval time.Duration
}

func NewPostgresLocker(retryInterval time.Duration) *PostgresLocker {
	if retryInterval <= 0 {
		retryInterval = 100 * time.Millisecond
	}
	return &PostgresLocker{retryInterval: retryInterval}
}

func (l *PostgresLocker) Acquire(ctx context.Context, db *sqlx.DB, lockID int64) error {
	ticker := time.NewTicker(l.retryInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for advisory lock: %w", ctx.Err())
		case <-ticker.C:
			var locked bool
			if err := db.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&locked); err != nil {
				return fmt.Errorf("query advisory lock: %w", err)
			}
			if locked {
				return nil
			}
		}
	}
}

func (l *PostgresLocker) Release(ctx context.Context, db *sqlx.DB, lockID int64) error {
	var released bool
	if err := db.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", lockID).Scan(&released); err != nil {
		return fmt.Errorf("query advisory unlock: %w", err)
	}
	if !released {
		return errors.New("advisory lock was not held")
	}
	return nil
}
