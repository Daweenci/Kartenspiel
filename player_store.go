package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type PlayerDB struct {
	ID           string
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	LastLogin    *time.Time
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

func createPlayer(username, password string) (string, error) {
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
	INSERT INTO players (username, password_hash)
	VALUES (?, ?)
	RETURNING id
	`

	var playerID string

	err = db.QueryRow(query, username, hashedPassword).Scan(&playerID)
	if err != nil {
		return "", fmt.Errorf("failed to create player: %w", err)
	}

	return playerID, nil
}

func verifyPlayerCredentials(username, password string) (string, error) {
	var playerID string
	var passwordHash string

	query := `SELECT id, password_hash FROM players WHERE username = ?`

	err := db.QueryRow(query, username).Scan(&playerID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	if !verifyPassword(passwordHash, password) {
		return "", fmt.Errorf("invalid credentials")
	}

	_, err = db.Exec(`UPDATE players SET last_login = CURRENT_TIMESTAMP WHERE id = ?`, playerID)
	if err != nil {
		log.Printf("Failed to update last login for player %s: %v", playerID, err)
	}

	return playerID, nil
}

func getPlayerByID(playerID string) (*PlayerDB, error) {
	var player PlayerDB

	query := `
	SELECT id, username, password_hash, created_at, last_login
	FROM players
	WHERE id = ?
	`

	err := db.QueryRow(query, playerID).Scan(
		&player.ID,
		&player.Username,
		&player.PasswordHash,
		&player.CreatedAt,
		&player.LastLogin,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("player not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &player, nil
}
