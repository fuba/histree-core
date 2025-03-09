// Package histree provides core functionality for shell history management
package histree

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Version information
const (
	Version = "v0.3.4"
)

// OutputFormat defines how history entries are formatted when displayed
type OutputFormat string

const (
	// FormatJSON outputs entries as JSON objects
	FormatJSON OutputFormat = "json"
	// FormatSimple outputs only the command
	FormatSimple OutputFormat = "simple"
	// FormatVerbose outputs entries with timestamp, directory and exit code
	FormatVerbose OutputFormat = "verbose"
)

// HistoryEntry represents a shell command history entry
type HistoryEntry struct {
	Command   string    `json:"command"`
	Directory string    `json:"directory"`
	Timestamp time.Time `json:"timestamp"`
	ExitCode  int       `json:"exit_code"`
	Hostname  string    `json:"hostname,omitempty"`
	ProcessID int       `json:"process_id,omitempty"` // The process ID of the shell that executed the command
}

// DB represents a histree database connection
type DB struct {
	*sql.DB
}

// OpenDB initializes and returns a new database connection
func OpenDB(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set PRAGMA for performance optimization
	if err := setPragmas(db); err != nil {
		return nil, err
	}

	// Create tables and indexes within a transaction
	if err := createSchema(db); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}

func setPragmas(db *sql.DB) error {
	_, err := db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA synchronous = NORMAL;
		PRAGMA temp_store = MEMORY;
		PRAGMA cache_size = -2000;
	`)
	if err != nil {
		return fmt.Errorf("failed to set pragmas: %w", err)
	}
	return nil
}

func createSchema(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create table with process information
	if err := createTable(tx); err != nil {
		return err
	}

	// Create indexes
	if err := createIndexes(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func createTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			command TEXT NOT NULL,
			directory TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			exit_code INTEGER NOT NULL,
			hostname TEXT NOT NULL,
			process_id INTEGER NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func createIndexes(tx *sql.Tx) error {
	queries := []string{
		`CREATE INDEX IF NOT EXISTS idx_history_directory ON history(directory)`,
		`CREATE INDEX IF NOT EXISTS idx_history_timestamp_directory ON history(timestamp, directory)`,
	}

	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}
	return nil
}

// AddEntry adds a new command history entry to the database
func (db *DB) AddEntry(entry *HistoryEntry) error {
	_, err := db.Exec(
		"INSERT INTO history (command, directory, timestamp, exit_code, hostname, process_id) VALUES (?, ?, ?, ?, ?, ?)",
		entry.Command,
		entry.Directory,
		entry.Timestamp,
		entry.ExitCode,
		entry.Hostname,
		entry.ProcessID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}
	return nil
}

// GetEntries retrieves command history entries from the database
func (db *DB) GetEntries(limit int, currentDir string) ([]HistoryEntry, error) {
	entries := make([]HistoryEntry, 0, limit)

	// Modified query to get the last N entries in chronological order
	query := `
		WITH recent_entries AS (
			SELECT command, directory, timestamp, exit_code, hostname, process_id
			FROM history 
			WHERE directory = ? OR directory LIKE ? || '/%'
			ORDER BY timestamp DESC
			LIMIT ?
		)
		SELECT * FROM recent_entries ORDER BY timestamp ASC
	`

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("PRAGMA page_size = 4096")
	if err != nil {
		return nil, fmt.Errorf("failed to set page size: %w", err)
	}

	// Pass currentDir twice for the two placeholders (exact match and LIKE pattern)
	rows, err := tx.Query(query, currentDir, currentDir, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry HistoryEntry
		err := rows.Scan(
			&entry.Command,
			&entry.Directory,
			&entry.Timestamp,
			&entry.ExitCode,
			&entry.Hostname,
			&entry.ProcessID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return entries, nil
}
