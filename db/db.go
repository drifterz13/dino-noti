package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Item struct {
	URL         string
	Name        string
	Price       string
	MatchedTerm string
	Timestamp   time.Time
}

func InitDB(databasePath string) (*sql.DB, error) {
	dbDir := filepath.Dir(databasePath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory %s: %w", dbDir, err)
	}
	fmt.Printf("Ensured database directory %s exists.\n", dbDir)

	db, err := sql.Open("sqlite3", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", databasePath, err)
	}

	// Create table if it doesn't exist
	createTableSQL :=
		`CREATE TABLE IF NOT EXISTS matched_items (
        url TEXT PRIMARY KEY,
        name TEXT,
        price TEXT,
        matched_term TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close() // Close connection before returning error
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	fmt.Println("Database initialized and table 'matched_items' checked/created.")
	return db, nil
}

// SaveItem saves a matched item to the database if its URL does not already exist.
func SaveItem(db *sql.DB, item Item) error {
	// Check if item already exists by URL
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM matched_items WHERE url = ?)", item.URL).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if item exists: %w", err)
	}

	if exists {
		fmt.Printf("Item already exists, skipping save: %s\n", item.URL)
		return nil // Item already exists, do nothing
	}

	// Insert the new item
	insertSQL :=
		`INSERT INTO matched_items (url, name, price, matched_term, timestamp) VALUES (?, ?, ?, ?, ?)`
	_, err = db.Exec(insertSQL, item.URL, item.Name, item.Price, item.MatchedTerm, item.Timestamp)
	if err != nil {
		return fmt.Errorf("failed to insert item %s: %w", item.URL, err)
	}

	fmt.Printf("Saved new matched item: %s\n", item.Name)
	return nil
}

// CloseDB closes the database connection.
func CloseDB(db *sql.DB) {
	if db != nil {
		db.Close()
		fmt.Println("Database connection closed.")
	}
}
