package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fuba/histree-core/pkg/histree"
)

func main() {
	version := flag.Bool("version", false, "Show version information")
	dbPath := flag.String("db", "", "Path to SQLite database (required)")
	action := flag.String("action", "", "Action to perform: add, get, or update-path")
	limit := flag.Int("limit", 100, "Number of entries to retrieve")
	currentDir := flag.String("dir", "", "Current directory for filtering entries")
	format := flag.String("format", string(histree.FormatSimple), "Output format: json, simple, or verbose")
	hostname := flag.String("hostname", "", "Hostname (required for add action)")
	processID := flag.Int("pid", 0, "Process ID (required for add action)")
	verbose := flag.Bool("v", false, "Show verbose output (same as -format verbose)")
	exitCode := flag.Int("exit", 0, "Exit code of the command")
	oldPath := flag.String("old-path", "", "Old directory path (required for update-path action)")
	newPath := flag.String("new-path", "", "New directory path (required for update-path action)")
	flag.Parse()

	if *version {
		fmt.Printf("histree %s\n", histree.Version)
		os.Exit(0)
	}

	if *dbPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -db parameter is required\n")
		flag.Usage()
		os.Exit(1)
	}

	// Override format if verbose flag is set
	if *verbose {
		*format = string(histree.FormatVerbose)
	}

	db, err := histree.OpenDB(*dbPath)
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
		if err := handleGet(db, *limit, *currentDir, histree.OutputFormat(*format)); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get entries: %v\n", err)
			os.Exit(1)
		}
		
	case "update-path":
		if *oldPath == "" || *newPath == "" {
			fmt.Fprintf(os.Stderr, "Error: both -old-path and -new-path parameters are required for update-path action\n")
			flag.Usage()
			os.Exit(1)
		}
		if err := handleUpdatePath(db, *oldPath, *newPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to update paths: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown action: %s\n", *action)
		os.Exit(1)
	}
}

func handleAdd(db *histree.DB, currentDir string, hostname string, processID int, exitCode int) error {
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

	entry := histree.HistoryEntry{
		Command:   cmd,
		Directory: dir,
		Timestamp: time.Now().UTC(),
		ExitCode:  exitCode,
		Hostname:  hostname,
		ProcessID: processID,
	}

	return db.AddEntry(&entry)
}

func handleGet(db *histree.DB, limit int, currentDir string, format histree.OutputFormat) error {
	entries, err := db.GetEntries(limit, currentDir)
	if err != nil {
		return err
	}

	return histree.WriteEntries(entries, os.Stdout, format)
}

func handleUpdatePath(db *histree.DB, oldPath, newPath string) error {
	// Convert to absolute paths if they aren't already
	if !filepath.IsAbs(oldPath) {
		absOldPath, err := filepath.Abs(oldPath)
		if err != nil {
			return fmt.Errorf("failed to convert old path to absolute path: %w", err)
		}
		oldPath = absOldPath
	}
	
	if !filepath.IsAbs(newPath) {
		absNewPath, err := filepath.Abs(newPath)
		if err != nil {
			return fmt.Errorf("failed to convert new path to absolute path: %w", err)
		}
		newPath = absNewPath
	}
	
	// Clean the paths to ensure consistent format
	oldPath = filepath.Clean(oldPath)
	newPath = filepath.Clean(newPath)
	
	// Update the paths in the database
	count, err := db.UpdatePaths(oldPath, newPath)
	if err != nil {
		return err
	}
	
	fmt.Printf("Updated %d entries: %s -> %s\n", count, oldPath, newPath)
	return nil
}
