package migrate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	gomigrate "github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
)

const defaultAdvisoryLockID int64 = 987654321

type Config struct {
	AutoMigrate bool
	LockTimeout time.Duration
	Path        string
}

type Locker interface {
	Acquire(ctx context.Context, db *sqlx.DB, lockID int64) error
	Release(ctx context.Context, db *sqlx.DB, lockID int64) error
}

type UpRunner interface {
	Up(databaseDSN, migrationsPath string) error
}

type Runner struct {
	locker Locker
	up     UpRunner
	lockID int64
}

func NewRunner() *Runner {
	return &Runner{
		locker: NewPostgresLocker(100 * time.Millisecond),
		up:     golangMigrateRunner{},
		lockID: defaultAdvisoryLockID,
	}
}

func (r *Runner) Run(ctx context.Context, cfg Config, db *sqlx.DB, databaseDSN string) error {
	if !cfg.AutoMigrate {
		log.Printf("startup migrations skipped: database.auto_migrate=false")
		return nil
	}
	if cfg.LockTimeout <= 0 {
		return errors.New("auto-migrate lock timeout must be > 0")
	}
	if cfg.Path == "" {
		return errors.New("migrations path must not be empty")
	}

	lockCtx, cancel := context.WithTimeout(ctx, cfg.LockTimeout)
	defer cancel()

	log.Printf("startup migrations: acquiring advisory lock")
	if err := r.locker.Acquire(lockCtx, db, r.lockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		releaseCtx, releaseCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer releaseCancel()
		if err := r.locker.Release(releaseCtx, db, r.lockID); err != nil {
			log.Printf("startup migrations: advisory lock release failed: %v", err)
		}
	}()

	log.Printf("startup migrations: running up migrations (path=%s)", cfg.Path)
	if err := r.up.Up(databaseDSN, cfg.Path); err != nil {
		if errors.Is(err, gomigrate.ErrNoChange) {
			log.Printf("startup migrations: no change")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Printf("startup migrations: complete")
	return nil
}
