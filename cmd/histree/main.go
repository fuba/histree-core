package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatReadable OutputFormat = "readable"
)

type HistoryEntry struct {
	Command      string    `json:"command"`
	Directory    string    `json:"directory"`
	Timestamp    time.Time `json:"timestamp"`
	SessionLabel string    `json:"session_label"`
}

func initDB(dbPath string) (*sql.DB, error) {
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

	return db, nil
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
			session_label TEXT NOT NULL
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

func addEntry(db *sql.DB, entry *HistoryEntry) error {
	_, err := db.Exec(
		"INSERT INTO history (command, directory, timestamp, session_label) VALUES (?, ?, ?, ?)",
		entry.Command,
		entry.Directory,
		entry.Timestamp,
		entry.SessionLabel,
	)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}
	return nil
}

func getEntries(db *sql.DB, limit int, currentDir string) ([]HistoryEntry, error) {
	entries := make([]HistoryEntry, 0, limit)

	query := `
		SELECT command, directory, timestamp, session_label
		FROM history 
		WHERE directory LIKE ? || '%'
		ORDER BY timestamp ASC
		LIMIT ?
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

	rows, err := tx.Query(query, currentDir, limit)
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
			&entry.SessionLabel,
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

func writeEntries(entries []HistoryEntry, w io.Writer, format OutputFormat) error {
	bufW := bufio.NewWriterSize(w, 8192)

	for _, entry := range entries {
		switch format {
		case FormatJSON:
			if err := json.NewEncoder(bufW).Encode(entry); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		case FormatReadable:
			command := entry.Command
			if strings.HasPrefix(command, "{") && strings.HasSuffix(command, "}") {
				command = fmt.Sprintf("%q", command)
			}

			if _, err := fmt.Fprintf(bufW, "%s [%s] (%s) %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.Directory,
				entry.SessionLabel,
				command); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		default:
			return fmt.Errorf("unknown output format: %s", format)
		}
	}

	if err := bufW.Flush(); err != nil {
		return fmt.Errorf("failed to flush buffer: %w", err)
	}
	return nil
}

func main() {
	dbPath := flag.String("db", "", "Path to SQLite database (required)")
	action := flag.String("action", "", "Action to perform: add or get")
	limit := flag.Int("limit", 100, "Number of entries to retrieve")
	currentDir := flag.String("dir", "", "Current directory for filtering entries")
	format := flag.String("format", string(FormatReadable), "Output format: json or readable")
	sessionLabel := flag.String("session", "", "Session label for command history (required for add action)")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -db parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	db, err := initDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch *action {
	case "add":
		if *sessionLabel == "" {
			fmt.Fprintf(os.Stderr, "Error: -session parameter is required for add action\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := handleAdd(db, *currentDir, *sessionLabel); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add entry: %v\n", err)
			os.Exit(1)
		}

	case "get":
		if err := handleGet(db, *limit, *currentDir, OutputFormat(*format)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get entries: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", *action)
		os.Exit(1)
	}
}

func handleAdd(db *sql.DB, currentDir string, sessionLabel string) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, os.Stdin); err != nil {
		return fmt.Errorf("failed to read command from stdin: %w", err)
	}
	cmd := strings.TrimRight(buf.String(), "\n")

	dir := currentDir
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	entry := HistoryEntry{
		Command:      cmd,
		Directory:    dir,
		Timestamp:    time.Now().UTC(),
		SessionLabel: sessionLabel,
	}

	return addEntry(db, &entry)
}

func handleGet(db *sql.DB, limit int, currentDir string, format OutputFormat) error {
	entries, err := getEntries(db, limit, currentDir)
	if err != nil {
		return err
	}

	return writeEntries(entries, os.Stdout, format)
}
