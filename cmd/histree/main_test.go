package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	dbPath := "./test_histree.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := createSchema(db); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}

	return db, cleanup
}

func TestAddEntry(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	entry := &HistoryEntry{
		Command:   "echo 'Hello, World!'",
		Directory: "/home/user",
		Timestamp: time.Now().UTC(),
		ExitCode:  0,
		Hostname:  "test-host",
		ProcessID: 12345,
	}

	if err := addEntry(db, entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	rows, err := db.Query("SELECT command, directory, timestamp, exit_code, hostname, process_id FROM history")
	if err != nil {
		t.Fatalf("Failed to query entries: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		count++
		var e HistoryEntry
		if err := rows.Scan(&e.Command, &e.Directory, &e.Timestamp, &e.ExitCode, &e.Hostname, &e.ProcessID); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		if e.Command != entry.Command ||
			e.Directory != entry.Directory ||
			e.ExitCode != entry.ExitCode ||
			e.Hostname != entry.Hostname ||
			e.ProcessID != entry.ProcessID {
			t.Errorf("Entry does not match: got %+v, want %+v", e, entry)
		}
	}

	if count != 1 {
		t.Errorf("Expected 1 entry, got %d", count)
	}
}

func TestGetEntries(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	entries := []HistoryEntry{
		{
			Command:   "echo 'Hello, World!'",
			Directory: "/home/user",
			Timestamp: time.Now().UTC(),
			ExitCode:  0,
			Hostname:  "test-host",
			ProcessID: 12345,
		},
		{
			Command:   "ls -la",
			Directory: "/home/user",
			Timestamp: time.Now().UTC(),
			ExitCode:  0,
			Hostname:  "test-host",
			ProcessID: 12345,
		},
	}

	for _, entry := range entries {
		if err := addEntry(db, &entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	gotEntries, err := getEntries(db, 10, "/home/user")
	if err != nil {
		t.Fatalf("Failed to get entries: %v", err)
	}

	if len(gotEntries) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), len(gotEntries))
	}

	for i, got := range gotEntries {
		want := entries[i]
		if got.Command != want.Command ||
			got.Directory != want.Directory ||
			got.ExitCode != want.ExitCode ||
			got.Hostname != want.Hostname ||
			got.ProcessID != want.ProcessID {
			t.Errorf("Entry %d does not match: got %+v, want %+v", i, got, want)
		}
	}
}
