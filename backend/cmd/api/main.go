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

	"basepro/backend/internal/audit"
	"basepro/backend/internal/auth"
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

	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSigningKey, time.Duration(cfg.Auth.AccessTokenTTLSeconds)*time.Second)
	authService := auth.NewService(
		auth.NewSQLRepository(database),
		audit.NewService(audit.NewSQLRepository(database)),
		jwtManager,
		time.Duration(cfg.Auth.AccessTokenTTLSeconds)*time.Second,
		time.Duration(cfg.Auth.RefreshTokenTTLSeconds)*time.Second,
	)

	srv := &http.Server{
		Addr: cfg.Server.Port,
		Handler: newRouter(AppDeps{
			DB:          database,
			Version:     version,
			AuthHandler: auth.NewHandler(authService),
			JWTManager:  jwtManager,
		}),
	}

	shutdownTimeout := time.Duration(cfg.Server.ShutdownTimeoutSeconds) * time.Second
	return runServer(ctx, srv, shutdownTimeout)
}

type cliFlags struct {
	fs               *flag.FlagSet
	configFile       string
	serverPort       string
	shutdownTimeout  int
	databaseDSN      string
	maxOpenConns     int
	maxIdleConns     int
	autoMigrate      bool
	authAccessTTL    int
	authRefreshTTL   int
	authSigningKey   string
	passwordHashCost int
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
	f.fs.IntVar(&f.authAccessTTL, "auth-access-ttl", 0, "access token TTL in seconds")
	f.fs.IntVar(&f.authRefreshTTL, "auth-refresh-ttl", 0, "refresh token TTL in seconds")
	f.fs.StringVar(&f.authSigningKey, "auth-jwt-signing-key", "", "JWT signing key")
	f.fs.IntVar(&f.passwordHashCost, "auth-password-hash-cost", 0, "bcrypt password hash cost")
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
		case "auth-access-ttl":
			overrides["auth.access_token_ttl_seconds"] = f.authAccessTTL
		case "auth-refresh-ttl":
			overrides["auth.refresh_token_ttl_seconds"] = f.authRefreshTTL
		case "auth-jwt-signing-key":
			overrides["auth.jwt_signing_key"] = f.authSigningKey
		case "auth-password-hash-cost":
			overrides["auth.password_hash_cost"] = f.passwordHashCost
		}
	})
	return overrides
}
