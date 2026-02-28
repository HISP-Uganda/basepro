package migrate

import (
	"fmt"
	"path/filepath"
	"strings"

	gomigrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type golangMigrateRunner struct{}

func (golangMigrateRunner) Up(databaseDSN, migrationsPath string) error {
	sourceURL, err := normalizeMigrationsPath(migrationsPath)
	if err != nil {
		return err
	}

	m, err := gomigrate.New(sourceURL, databaseDSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil {
		return err
	}
	return nil
}

func normalizeMigrationsPath(path string) (string, error) {
	if strings.Contains(path, "://") {
		return path, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve migrations path: %w", err)
	}
	return "file://" + filepath.ToSlash(absPath), nil
}

func closeMigrator(m *gomigrate.Migrate) {
	sourceErr, dbErr := m.Close()
	_ = sourceErr
	_ = dbErr
}
