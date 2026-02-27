package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"basepro/backend/internal/config"
	"basepro/backend/internal/migrateutil"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate [up|down|create]")
	}

	command := os.Args[1]
	switch command {
	case "up":
		runUpDown(true)
	case "down":
		runUpDown(false)
	case "create":
		runCreate()
	default:
		log.Fatalf("unknown command %q", command)
	}
}

func runUpDown(isUp bool) {
	fs := flag.NewFlagSet("migrate", flag.ExitOnError)
	configFile := fs.String("config", "", "path to config file")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	if _, err := config.Load(config.Options{ConfigFile: *configFile}); err != nil {
		log.Fatalf("load config: %v", err)
	}

	cfg := config.Get()
	if isUp {
		if err := migrateutil.Up(cfg.Database.DSN, "./migrations"); err != nil {
			log.Fatalf("migrate up failed: %v", err)
		}
		fmt.Println("migrations applied")
		return
	}

	if err := migrateutil.DownOne(cfg.Database.DSN, "./migrations"); err != nil {
		log.Fatalf("migrate down failed: %v", err)
	}
	fmt.Println("one migration rolled back")
}

func runCreate() {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	name := fs.String("name", "", "migration name")
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	up, down, err := migrateutil.CreatePair("./migrations", *name)
	if err != nil {
		log.Fatalf("create migration files: %v", err)
	}

	fmt.Printf("created %s\n", up)
	fmt.Printf("created %s\n", down)
}
