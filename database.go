package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func initDB() error {
	dbPath := "game.db" // SQLite database file

	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	if err = loadSchema(); err != nil {
		return fmt.Errorf("failed to load schema: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

func loadSchema() error {
	schemaFile, err := os.Open("schema.sql")
	if err != nil {
		return fmt.Errorf("failed to open schema.sql: %v", err)
	}
	defer schemaFile.Close()

	schemaBytes, err := io.ReadAll(schemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema.sql: %v", err)
	}

	schema := string(schemaBytes)

	// Remove comments and split properly
	lines := strings.Split(schema, "\n")
	var cleanedLines []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "--") {
			cleanedLines = append(cleanedLines, line)
		}
	}

	cleanedSchema := strings.Join(cleanedLines, " ")
	statements := strings.Split(cleanedSchema, ";")

	for _, statement := range statements {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}

		log.Printf("Executing SQL: %s", statement) // Debug output
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("failed to execute statement '%s': %v", statement, err)
		}
	}

	return nil
}

type PlayerDB struct {
	ID           string     `db:"id"`
	Username     string     `db:"username"`
	PasswordHash string     `db:"password_hash"`
	CreatedAt    time.Time  `db:"created_at"`
	LastLogin    *time.Time `db:"last_login"`
	IsOnline     bool       `db:"is_online"`
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func registerPlayer(username, password string) (string, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}

	// SQLite doesn't support RETURNING, so we need to do this differently
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO players (username, password_hash) VALUES (?, ?)`
	result, err := tx.Exec(query, username, hashedPassword)
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: players.username" {
			return "", fmt.Errorf("username already exists")
		}
		return "", fmt.Errorf("failed to create player: %v", err)
	}

	// Get the last inserted row ID and then fetch the actual ID
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("failed to get last insert ID: %v", err)
	}

	var playerID string
	err = tx.QueryRow(`SELECT id FROM players WHERE rowid = ?`, lastInsertID).Scan(&playerID)
	if err != nil {
		return "", fmt.Errorf("failed to get player ID: %v", err)
	}

	if err = tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return playerID, nil
}

func authenticatePlayer(username, password string) (string, error) {
	var playerDB PlayerDB
	query := `SELECT id, password_hash FROM players WHERE username = ?`
	err := db.QueryRow(query, username).Scan(&playerDB.ID, &playerDB.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", fmt.Errorf("database error: %v", err)
	}

	if !verifyPassword(playerDB.PasswordHash, password) {
		return "", fmt.Errorf("invalid credentials")
	}

	updateQuery := `UPDATE players SET last_login = CURRENT_TIMESTAMP, is_online = TRUE WHERE id = ?`
	_, err = db.Exec(updateQuery, playerDB.ID)
	if err != nil {
		log.Printf("Failed to update last login for player %s: %v", playerDB.ID, err)
	}

	return playerDB.ID, nil
}

func getPlayerByID(playerID string) (*PlayerDB, error) {
	var player PlayerDB
	query := `SELECT id, username, password_hash, created_at, last_login, is_online 
			  FROM players WHERE id = ?`
	err := db.QueryRow(query, playerID).Scan(
		&player.ID, &player.Username, &player.PasswordHash,
		&player.CreatedAt, &player.LastLogin, &player.IsOnline)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}
	return &player, nil
}

func setPlayerOnlineStatus(playerID string, isOnline bool) error {
	query := `UPDATE players SET is_online = ? WHERE id = ?`
	_, err := db.Exec(query, isOnline, playerID)
	if err != nil {
		return fmt.Errorf("failed to update online status: %v", err)
	}
	return nil
}

func closeDB() {
	if db != nil {
		db.Close()
	}
}
