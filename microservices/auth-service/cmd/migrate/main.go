package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	// Dapatkan koneksi string dari env variable
	dbURL := os.Getenv("AUTH_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://user:pass@localhost:5432/auth_db?sslmode=disable"
	}

	// Buka koneksi database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
	defer db.Close()

	// Setup driver migrasi
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	// Buat instance migrasi
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations", // Path ke folder migrations
		"postgres", driver)
	if err != nil {
		log.Fatalf("failed to create migration instance: %v", err)
	}

	// Eksekusi migrasi berdasarkan command
	switch os.Getenv("MIGRATE_CMD") {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to apply migrations: %v", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("failed to revert migrations: %v", err)
		}
		log.Println("Migrations reverted successfully")
	default:
		log.Fatal("Set MIGRATE_CMD environment variable to 'up' or 'down'")
	}
}
