package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() error {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("DB_PORT")
		if port == "" {
			port = "5432"
		}
		user := os.Getenv("DB_USER")
		if user == "" {
			user = "postgres"
		}
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")
		if dbname == "" {
			dbname = "mydb"
		}

		dbURL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)
	}

	var err error
	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database connection established successfully")
	return nil
}

func createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS players (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			username VARCHAR(50) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_login TIMESTAMP,
			is_online BOOLEAN DEFAULT FALSE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_players_username ON players(username)`,
		`CREATE INDEX IF NOT EXISTS idx_players_online ON players(is_online)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query '%s': %v", query, err)
		}
	}

	return nil
}

type PlayerDB struct {
	ID           string    `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	LastLogin    *time.Time `db:"last_login"`
	IsOnline     bool      `db:"is_online"`
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

	var playerID string
	query := `INSERT INTO players (username, password_hash) VALUES ($1, $2) RETURNING id`
	err = db.QueryRow(query, username, hashedPassword).Scan(&playerID)
	if err != nil {
		if err.Error() == `pq: duplicate key value violates unique constraint "players_username_key"` {
			return "", fmt.Errorf("username already exists")
		}
		return "", fmt.Errorf("failed to create player: %v", err)
	}

	return playerID, nil
}

func authenticatePlayer(username, password string) (string, error) {
	var playerDB PlayerDB
	query := `SELECT id, password_hash FROM players WHERE username = $1`
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

	updateQuery := `UPDATE players SET last_login = CURRENT_TIMESTAMP, is_online = TRUE WHERE id = $1`
	_, err = db.Exec(updateQuery, playerDB.ID)
	if err != nil {
		log.Printf("Failed to update last login for player %s: %v", playerDB.ID, err)
	}

	return playerDB.ID, nil
}

func getPlayerByID(playerID string) (*PlayerDB, error) {
	var player PlayerDB
	query := `SELECT id, username, password_hash, created_at, last_login, is_online 
			  FROM players WHERE id = $1`
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
	query := `UPDATE players SET is_online = $1 WHERE id = $2`
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