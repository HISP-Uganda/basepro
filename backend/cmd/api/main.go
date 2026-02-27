package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"basepro/backend/internal/config"
	"basepro/backend/internal/db"
	"basepro/backend/internal/migrateutil"
)

const version = "0.1.0"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	flags := newFlags()
	if err := flags.fs.Parse(os.Args[1:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	_, err := config.Load(config.Options{
		ConfigFile: flags.configFile,
		Overrides:  flags.overrides(),
		Watch:      true,
	})
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	cfg := config.Get()
	database, err := db.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			log.Printf("database close error: %v", closeErr)
		}
	}()

	if cfg.Database.AutoMigrate {
		if err := migrateutil.Up(cfg.Database.DSN, "./migrations"); err != nil {
			return fmt.Errorf("auto-migrate: %w", err)
		}
	}

	srv := &http.Server{
		Addr:    cfg.Server.Port,
		Handler: newRouter(AppDeps{DB: database, Version: version}),
	}

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeoutSeconds) * time.Second
	return runServer(ctx, srv, shutdownTimeout)
}

type cliFlags struct {
	fs              *flag.FlagSet
	configFile      string
	serverPort      string
	shutdownTimeout int
	databaseDSN     string
	maxOpenConns    int
	maxIdleConns    int
	autoMigrate     bool
}

func newFlags() *cliFlags {
	f := &cliFlags{fs: flag.NewFlagSet(os.Args[0], flag.ContinueOnError)}
	f.fs.StringVar(&f.configFile, "config", "", "path to config file")
	f.fs.StringVar(&f.serverPort, "server-port", "", "server listen address")
	f.fs.IntVar(&f.shutdownTimeout, "shutdown-timeout", 0, "shutdown timeout in seconds")
	f.fs.StringVar(&f.databaseDSN, "database-dsn", "", "database DSN")
	f.fs.IntVar(&f.maxOpenConns, "database-max-open-conns", 0, "max open DB connections")
	f.fs.IntVar(&f.maxIdleConns, "database-max-idle-conns", 0, "max idle DB connections")
	f.fs.BoolVar(&f.autoMigrate, "database-auto-migrate", false, "auto-run migrations on startup")
	return f
}

func (f *cliFlags) overrides() map[string]any {
	overrides := make(map[string]any)
	f.fs.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "server-port":
			overrides["server.port"] = f.serverPort
		case "shutdown-timeout":
			overrides["server.shutdown_timeout_seconds"] = f.shutdownTimeout
		case "database-dsn":
			overrides["database.dsn"] = f.databaseDSN
		case "database-max-open-conns":
			overrides["database.max_open_conns"] = f.maxOpenConns
		case "database-max-idle-conns":
			overrides["database.max_idle_conns"] = f.maxIdleConns
		case "database-auto-migrate":
			overrides["database.auto_migrate"] = f.autoMigrate
		}
	})
	return overrides
}
