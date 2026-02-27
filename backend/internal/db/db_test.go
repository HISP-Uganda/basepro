package db

import (
	"context"
	"os"
	"testing"
	"time"

	"basepro/backend/internal/config"
)

func TestOpenIntegration(t *testing.T) {
	dsn := os.Getenv("BASEPRO_TEST_DSN")
	if dsn == "" {
		t.Skip("BASEPRO_TEST_DSN is not set")
	}

	cfg := config.Config{}
	cfg.Server.Port = ":8080"
	cfg.Server.ShutdownTimeoutSeconds = 10
	cfg.Database.DSN = dsn
	cfg.Database.MaxOpenConns = 2
	cfg.Database.MaxIdleConns = 1

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := Open(ctx, cfg)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
}
