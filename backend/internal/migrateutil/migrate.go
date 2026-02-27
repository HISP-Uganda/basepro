package migrateutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var migrationFileRE = regexp.MustCompile(`^(\d{6})_.*\.(up|down)\.sql$`)

func Up(databaseDSN, migrationsDir string) error {
	m, err := newMigrator(databaseDSN, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

func DownOne(databaseDSN, migrationsDir string) error {
	m, err := newMigrator(databaseDSN, migrationsDir)
	if err != nil {
		return err
	}
	defer closeMigrator(m)

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

func CreatePair(migrationsDir, name string) (string, string, error) {
	if name == "" {
		return "", "", errors.New("migration name must not be empty")
	}

	if err := os.MkdirAll(migrationsDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create migrations dir: %w", err)
	}

	next, err := nextSequence(migrationsDir)
	if err != nil {
		return "", "", err
	}

	prefix := fmt.Sprintf("%06d_%s", next, sanitizeName(name))
	upFile := filepath.Join(migrationsDir, prefix+".up.sql")
	downFile := filepath.Join(migrationsDir, prefix+".down.sql")

	stamp := time.Now().UTC().Format(time.RFC3339)
	content := fmt.Sprintf("-- created at %s\n", stamp)

	if err := os.WriteFile(upFile, []byte(content), 0o644); err != nil {
		return "", "", fmt.Errorf("write up migration: %w", err)
	}
	if err := os.WriteFile(downFile, []byte(content), 0o644); err != nil {
		return "", "", fmt.Errorf("write down migration: %w", err)
	}

	return upFile, downFile, nil
}

func newMigrator(databaseDSN, migrationsDir string) (*migrate.Migrate, error) {
	if databaseDSN == "" {
		return nil, errors.New("database DSN must not be empty")
	}

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("resolve migrations dir: %w", err)
	}

	sourceURL := "file://" + absDir
	m, err := migrate.New(sourceURL, databaseDSN)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return m, nil
}

func closeMigrator(m *migrate.Migrate) {
	srcErr, dbErr := m.Close()
	if srcErr != nil {
		fmt.Fprintf(os.Stderr, "close migration source error: %v\n", srcErr)
	}
	if dbErr != nil {
		fmt.Fprintf(os.Stderr, "close migration db error: %v\n", dbErr)
	}
}

func nextSequence(migrationsDir string) (int, error) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return 0, fmt.Errorf("read migrations dir: %w", err)
	}

	var sequences []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := migrationFileRE.FindStringSubmatch(entry.Name())
		if len(matches) < 2 {
			continue
		}
		seq, convErr := strconv.Atoi(matches[1])
		if convErr != nil {
			continue
		}
		sequences = append(sequences, seq)
	}

	if len(sequences) == 0 {
		return 1, nil
	}

	sort.Ints(sequences)
	return sequences[len(sequences)-1] + 1, nil
}

func sanitizeName(name string) string {
	name = filepath.Clean(name)
	name = filepath.Base(name)
	name = regexp.MustCompile(`[^a-zA-Z0-9_]+`).ReplaceAllString(name, "_")
	return name
}
