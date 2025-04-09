package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fuba/histree-core/pkg/histree"
)

func setupTestDB(t *testing.T) (*histree.DB, func()) {
	dbPath := "./test_histree.db"
	db, err := histree.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
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

	entry := &histree.HistoryEntry{
		Command:   "echo 'Hello, World!'",
		Directory: "/home/user",
		Timestamp: time.Now().UTC(),
		ExitCode:  0,
		Hostname:  "test-host",
		ProcessID: 12345,
	}

	if err := db.AddEntry(entry); err != nil {
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
		var e histree.HistoryEntry
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

	entries := []histree.HistoryEntry{
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
		if err := db.AddEntry(&entry); err != nil {
			t.Fatalf("Failed to add entry: %v", err)
		}
	}

	gotEntries, err := db.GetEntries(10, "/home/user")
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

// TestFormatVerboseWithTimezone tests that the FormatVerbose output
// correctly converts UTC timestamps to local timezone
func TestFormatVerboseWithTimezone(t *testing.T) {
	// Create a test entry with a fixed UTC time
	fixedUTCTime := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)
	entry := histree.HistoryEntry{
		Command:   "echo 'timezone test'",
		Directory: "/home/user/test",
		Timestamp: fixedUTCTime,
		ExitCode:  0,
		Hostname:  "test-host",
		ProcessID: 12345,
	}

	// Test output with verbose format
	var buf bytes.Buffer
	if err := histree.WriteEntries([]histree.HistoryEntry{entry}, &buf, histree.FormatVerbose); err != nil {
		t.Fatalf("Failed to write entries: %v", err)
	}

	outputStr := buf.String()

	// Expected output should have the local time, not UTC
	localTime := fixedUTCTime.Local()
	expectedTimePrefix := localTime.Format("2006-01-02T15:04:05")
	if !strings.Contains(outputStr, expectedTimePrefix) {
		t.Errorf("Expected output to contain local time %s, got: %s", expectedTimePrefix, outputStr)
	}

	// The output should NOT contain the UTC time
	utcTimePrefix := fixedUTCTime.Format("2006-01-02T15:04:05")
	if utcTimePrefix != expectedTimePrefix && strings.Contains(outputStr, utcTimePrefix) {
		t.Errorf("Output should not contain UTC time %s, got: %s", utcTimePrefix, outputStr)
	}
}

// TestFormatVerboseWithSpecificTimezones tests the timestamp display in different timezones
func TestFormatVerboseWithSpecificTimezones(t *testing.T) {
	// Save original timezone
	originalTZ := os.Getenv("TZ")
	defer os.Setenv("TZ", originalTZ)

	// Test with a few different timezones
	testTimezones := []struct {
		tz       string
		expected string // Expected hour part of the output (varies by timezone)
	}{
		{"UTC", "12:30"},
		{"America/New_York", "08:30"}, // UTC-4 (might vary with DST)
		{"Asia/Tokyo", "21:30"},       // UTC+9
	}

	// Fixed UTC time for testing
	fixedUTCTime := time.Date(2023, 5, 15, 12, 30, 0, 0, time.UTC)

	for _, tc := range testTimezones {
		t.Run("Timezone_"+tc.tz, func(t *testing.T) {
			// Set the timezone for this test
			os.Setenv("TZ", tc.tz)
			loc, err := time.LoadLocation(tc.tz)
			if err != nil {
				t.Fatalf("Failed to load location for %s: %v", tc.tz, err)
			}
			time.Local = loc

			entry := histree.HistoryEntry{
				Command:   "echo 'timezone test'",
				Directory: "/home/user/test",
				Timestamp: fixedUTCTime,
				ExitCode:  0,
				Hostname:  "test-host",
				ProcessID: 12345,
			}

			var buf bytes.Buffer
			if err := histree.WriteEntries([]histree.HistoryEntry{entry}, &buf, histree.FormatVerbose); err != nil {
				t.Fatalf("Failed to write entries in timezone %s: %v", tc.tz, err)
			}

			outputStr := buf.String()

			// Check if the output contains the expected time format in the current timezone without timezone offset
			if !strings.Contains(outputStr, tc.expected) {
				t.Errorf("In timezone %s: Expected time containing %s, got: %s",
					tc.tz, tc.expected, outputStr)
			}
		})
	}
}

// TestUpdatePaths tests the UpdatePaths functionality
func TestUpdatePaths(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Define test paths
	oldPath := "/home/user/oldpath"
	newPath := "/home/user/newpath"
	
	// Create test entries with different paths
	entries := []histree.HistoryEntry{
		{
			Command:   "cd /home/user/oldpath",
			Directory: oldPath,
			Timestamp: time.Now().UTC(),
			ExitCode:  0,
			Hostname:  "test-host",
			ProcessID: 12345,
		},
		{
			Command:   "ls -la",
			Directory: oldPath + "/subdir",
			Timestamp: time.Now().UTC(),
			ExitCode:  0,
			Hostname:  "test-host",
			ProcessID: 12345,
		},
		{
			Command:   "echo 'unrelated'",
			Directory: "/tmp",
			Timestamp: time.Now().UTC(),
			ExitCode:  0,
			Hostname:  "test-host",
			ProcessID: 12345,
		},
	}

	// Add test entries
	for _, entry := range entries {
		if err := db.AddEntry(&entry); err != nil {
			t.Fatalf("Failed to add test entry: %v", err)
		}
	}

	// Update paths
	count, err := db.UpdatePaths(oldPath, newPath)
	if err != nil {
		t.Fatalf("Failed to update paths: %v", err)
	}

	// Should have updated 2 entries (main path and subdirectory)
	if count != 2 {
		t.Errorf("Expected 2 entries to be updated, got %d", count)
	}

	// Verify the updates
	rows, err := db.Query("SELECT directory FROM history ORDER BY id")
	if err != nil {
		t.Fatalf("Failed to query entries: %v", err)
	}
	defer rows.Close()

	var updatedDirs []string
	for rows.Next() {
		var dir string
		if err := rows.Scan(&dir); err != nil {
			t.Fatalf("Failed to scan row: %v", err)
		}
		updatedDirs = append(updatedDirs, dir)
	}

	// Check the expected path changes
	expectedDirs := []string{
		newPath,                // oldPath should now be newPath
		newPath + "/subdir",    // oldPath/subdir should now be newPath/subdir
		"/tmp",                 // Unrelated path should remain unchanged
	}

	if len(updatedDirs) != len(expectedDirs) {
		t.Fatalf("Expected %d entries, got %d", len(expectedDirs), len(updatedDirs))
	}

	for i, expected := range expectedDirs {
		if updatedDirs[i] != expected {
			t.Errorf("Entry %d: expected directory %q, got %q", i, expected, updatedDirs[i])
		}
	}
}
