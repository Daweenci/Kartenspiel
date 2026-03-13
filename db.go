package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() error {
	dbPath := "game.db"

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// SQLite works best with a single connection
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	var version string
	if err := db.QueryRow("SELECT sqlite_version()").Scan(&version); err != nil {
		return fmt.Errorf("failed to get sqlite version: %w", err)
	}

	log.Println("SQLite version:", version)

	// Recommended SQLite settings for servers
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return fmt.Errorf("failed to enable WAL: %w", err)
	}

	if _, err := db.Exec(`PRAGMA synchronous=NORMAL;`); err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}

	if err := loadSchema(); err != nil {
		return err
	}

	log.Println("Database connection established successfully")

	return nil
}

func loadSchema() error {
	schemaBytes, err := os.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %w", err)
	}

	_, err = db.Exec(string(schemaBytes))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	return nil
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}
