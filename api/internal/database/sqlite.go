package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func NewConnection(dbPath string) (*sql.DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode and foreign keys for better performance
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to execute %s: %w", pragma, err)
		}
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS promo_codes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		code TEXT UNIQUE NOT NULL,
		reward_type TEXT NOT NULL,
		reward_amount REAL NOT NULL,
		reward_description_fr TEXT,
		reward_description_en TEXT,
		currency TEXT DEFAULT 'FCFA',
		status TEXT DEFAULT 'active',
		max_uses INTEGER DEFAULT 1,
		current_uses INTEGER DEFAULT 0,
		valid_from DATETIME,
		valid_until DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		transaction_id TEXT UNIQUE NOT NULL,
		promo_code_id INTEGER,
		phone_number TEXT NOT NULL,
		promo_code_input TEXT NOT NULL,
		status TEXT NOT NULL,
		error_code TEXT,
		failure_reason TEXT,
		reward_type TEXT,
		reward_amount REAL,
		reward_delivered INTEGER DEFAULT 0,
		language TEXT DEFAULT 'fr',
		whatsapp_message_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		processed_at DATETIME,
		FOREIGN KEY (promo_code_id) REFERENCES promo_codes(id)
	);

	CREATE TABLE IF NOT EXISTS rate_limits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		phone_number TEXT NOT NULL,
		window_start DATETIME NOT NULL,
		request_count INTEGER DEFAULT 1,
		window_type TEXT NOT NULL,
		UNIQUE(phone_number, window_type)
	);

	CREATE INDEX IF NOT EXISTS idx_promo_code ON promo_codes(code);
	CREATE INDEX IF NOT EXISTS idx_transactions_phone ON transactions(phone_number);
	CREATE INDEX IF NOT EXISTS idx_transactions_txid ON transactions(transaction_id);
	CREATE INDEX IF NOT EXISTS idx_rate_limits_phone ON rate_limits(phone_number);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func SeedTestData(db *sql.DB) error {
	seeds := `
	INSERT OR IGNORE INTO promo_codes (code, reward_type, reward_amount, reward_description_fr, reward_description_en, status, max_uses) VALUES
		('VALID100', 'data', 1, '1GB de data mobile', '1GB of mobile data', 'active', 100),
		('VALID200', 'credit', 500, '500 FCFA de credit', '500 FCFA credit', 'active', 100),
		('EXPIRED01', 'data', 0.5, '500MB de data', '500MB of data', 'expired', 1),
		('USED001', 'credit', 100, '100 FCFA de credit', '100 FCFA credit', 'used', 1),
		('INVALID', 'none', 0, '', '', 'invalid', 0);
	`

	_, err := db.Exec(seeds)
	if err != nil {
		return fmt.Errorf("failed to seed test data: %w", err)
	}

	return nil
}
