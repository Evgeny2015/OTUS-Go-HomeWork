package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func main() {
	var (
		command   = flag.String("command", "up", "goose command: up, down, status, create, etc.")
		dir       = flag.String("dir", "./migrations", "directory with migration files")
		dsn       = flag.String("dsn", "", "database connection string")
		tableName = flag.String("table", "goose_migrations", "migrations table name")
	)
	flag.Parse()

	if *dsn == "" {
		// Try to get DSN from environment or use default
		envDSN := os.Getenv("CALENDAR_STORAGE_DSN")
		if envDSN == "" {
			envDSN = "host=localhost port=5432 user=postgres password=postgres dbname=calendar sslmode=disable"
		}
		*dsn = envDSN
	}

	// Set goose table name
	goose.SetTableName(*tableName)

	// Open database connection
	db, err := goose.OpenDBWithDriver("postgres", *dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run goose command
	if err := goose.Run(*command, db, *dir, flag.Args()...); err != nil {
		log.Fatalf("Failed to run goose command %q: %v", *command, err)
	}

	fmt.Printf("Goose command %q completed successfully\n", *command)
}
