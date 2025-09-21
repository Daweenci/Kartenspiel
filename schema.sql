-- schema.sql - Database schema for the game

CREATE TABLE IF NOT EXISTS players (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    is_online BOOLEAN DEFAULT FALSE
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_online ON players(is_online);