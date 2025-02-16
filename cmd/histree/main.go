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
	FormatJSON    OutputFormat = "json"
	FormatSimple  OutputFormat = "simple"
	FormatVerbose OutputFormat = "verbose"
)

type HistoryEntry struct {
	Command   string    `json:"command"`
	Directory string    `json:"directory"`
	Timestamp time.Time `json:"timestamp"`
	ExitCode  int       `json:"exit_code"`
	Hostname  string    `json:"hostname,omitempty"`
	ProcessID int       `json:"process_id,omitempty"` // The process ID of the shell that executed the command
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

func addEntry(db *sql.DB, entry *HistoryEntry) error {
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

func getEntries(db *sql.DB, limit int, currentDir string) ([]HistoryEntry, error) {
	entries := make([]HistoryEntry, 0, limit)

	query := `
		SELECT command, directory, timestamp, exit_code, hostname, process_id
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

func writeEntries(entries []HistoryEntry, w io.Writer, format OutputFormat) error {
	bufW := bufio.NewWriterSize(w, 8192)

	for _, entry := range entries {
		switch format {
		case FormatJSON:
			if err := json.NewEncoder(bufW).Encode(entry); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		case FormatSimple:
			if _, err := fmt.Fprintf(bufW, "%s\n", entry.Command); err != nil {
				return fmt.Errorf("failed to write entry: %w", err)
			}
		case FormatVerbose:
			command := entry.Command
			if strings.HasPrefix(command, "{") && strings.HasSuffix(command, "}") {
				command = fmt.Sprintf("%q", command)
			}

			exitStatus := ""
			if entry.ExitCode != 0 {
				exitStatus = fmt.Sprintf(" [%d]", entry.ExitCode)
			}

			if _, err := fmt.Fprintf(bufW, "%s [%s]%s %s\n",
				entry.Timestamp.Format(time.RFC3339),
				entry.Directory,
				exitStatus,
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
	format := flag.String("format", string(FormatSimple), "Output format: json, simple, or verbose")
	hostname := flag.String("hostname", "", "Hostname (required for add action)")
	processID := flag.Int("pid", 0, "Process ID (required for add action)")
	verbose := flag.Bool("v", false, "Show verbose output (same as -format verbose)")
	exitCode := flag.Int("exit", 0, "Exit code of the command")
	flag.Parse()

	if *dbPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -db parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Override format if verbose flag is set
	if *verbose {
		*format = string(FormatVerbose)
	}

	db, err := initDB(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch *action {
	case "add":
		if *hostname == "" {
			fmt.Fprintf(os.Stderr, "Error: -hostname parameter is required for add action\n")
			flag.Usage()
			os.Exit(1)
		}
		if *processID == 0 {
			fmt.Fprintf(os.Stderr, "Error: -pid parameter is required for add action\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := handleAdd(db, *currentDir, *hostname, *processID, *exitCode); err != nil {
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

func handleAdd(db *sql.DB, currentDir string, hostname string, processID int, exitCode int) error {
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
		Command:   cmd,
		Directory: dir,
		Timestamp: time.Now().UTC(),
		ExitCode:  exitCode,
		Hostname:  hostname,
		ProcessID: processID,
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
