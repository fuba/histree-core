package histree

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasSufficientDiskSpace_MemoryDB(t *testing.T) {
	// In-memory database should always return true (skip disk check)
	hasSpace, err := hasSufficientDiskSpace(":memory:")
	if err != nil {
		t.Fatalf("unexpected error for :memory: db: %v", err)
	}
	if !hasSpace {
		t.Error("expected true for :memory: database")
	}
}

func TestHasSufficientDiskSpace_EmptyPath(t *testing.T) {
	// Empty path should always return true (skip disk check)
	hasSpace, err := hasSufficientDiskSpace("")
	if err != nil {
		t.Fatalf("unexpected error for empty path: %v", err)
	}
	if !hasSpace {
		t.Error("expected true for empty path")
	}
}

func TestHasSufficientDiskSpace_NonExistentDirectory(t *testing.T) {
	// Non-existent directory should find parent and check, or return true
	tmpDir := t.TempDir()
	nonExistentPath := filepath.Join(tmpDir, "does", "not", "exist", "test.db")

	hasSpace, err := hasSufficientDiskSpace(nonExistentPath)
	if err != nil {
		t.Fatalf("unexpected error for non-existent directory: %v", err)
	}
	// Should find tmpDir as parent and check disk space there
	// Since tmpDir exists and likely has space, should return true
	if !hasSpace {
		t.Error("expected true for non-existent directory with existing parent")
	}
}

func TestHasSufficientDiskSpace_ExistingDirectory(t *testing.T) {
	// Existing directory should work normally
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	hasSpace, err := hasSufficientDiskSpace(dbPath)
	if err != nil {
		t.Fatalf("unexpected error for existing directory: %v", err)
	}
	// tmpDir is a real directory with space, should return true
	if !hasSpace {
		t.Error("expected true for existing directory with space")
	}
}

func TestHasSufficientDiskSpace_Symlink(t *testing.T) {
	// Test symlink resolution
	tmpDir := t.TempDir()
	realDir := filepath.Join(tmpDir, "real")
	linkDir := filepath.Join(tmpDir, "link")

	if err := os.Mkdir(realDir, 0755); err != nil {
		t.Fatalf("failed to create real directory: %v", err)
	}

	if err := os.Symlink(realDir, linkDir); err != nil {
		t.Skipf("symlinks not supported on this system: %v", err)
	}

	dbPath := filepath.Join(linkDir, "test.db")
	hasSpace, err := hasSufficientDiskSpace(dbPath)
	if err != nil {
		t.Fatalf("unexpected error for symlink path: %v", err)
	}
	if !hasSpace {
		t.Error("expected true for symlink path")
	}
}

func TestHasSufficientDiskSpace_PathIsFile(t *testing.T) {
	// If the "directory" is actually a file, should return error
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "somefile")

	// Create a file instead of a directory
	if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to use this file as if it were a directory containing a db
	dbPath := filepath.Join(filePath, "test.db")
	hasSpace, err := hasSufficientDiskSpace(dbPath)

	// filepath.Dir(dbPath) == filePath, which is a file not a directory
	if err == nil {
		t.Error("expected error when parent path is a file, not a directory")
	}
	if hasSpace {
		t.Error("expected false when parent path is a file")
	}
}

func TestFindExistingParent(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		dir      string
		expected string
	}{
		{
			name:     "non-existent nested path",
			dir:      filepath.Join(tmpDir, "a", "b", "c"),
			expected: tmpDir,
		},
		{
			name:     "existing directory",
			dir:      tmpDir,
			expected: filepath.Dir(tmpDir), // Parent of tmpDir
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := findExistingParent(tc.dir)
			if tc.expected != "" && result == "" {
				t.Errorf("expected to find parent, got empty string")
			}
			if result != "" {
				// Verify the returned path exists
				if _, err := os.Stat(result); err != nil {
					t.Errorf("returned parent %q does not exist: %v", result, err)
				}
			}
		})
	}
}

func TestFindExistingParent_RootPath(t *testing.T) {
	// Testing with absolute non-existent path
	result := findExistingParent("/nonexistent/deeply/nested/path")
	// Should eventually find "/" on Unix systems
	if result == "" {
		t.Log("findExistingParent returned empty for root-relative path (may be expected on some systems)")
	} else {
		if _, err := os.Stat(result); err != nil {
			t.Errorf("returned parent %q does not exist: %v", result, err)
		}
	}
}

func TestSetDiskSpaceChecker(t *testing.T) {
	// Test that SetDiskSpaceChecker works correctly
	originalChecker := diskSpaceCheckerFn

	// Set custom checker
	customCalled := false
	SetDiskSpaceChecker(func(path string) (bool, error) {
		customCalled = true
		return false, nil
	})

	// Verify custom checker is called
	_, _ = checkDiskSpace("/some/path")
	if !customCalled {
		t.Error("custom checker was not called")
	}

	// Reset to nil (should restore default)
	SetDiskSpaceChecker(nil)

	// Verify default is restored
	diskSpaceCheckerMu.RLock()
	currentChecker := diskSpaceCheckerFn
	diskSpaceCheckerMu.RUnlock()

	// Can't directly compare functions, but we can check it's not nil
	if currentChecker == nil {
		t.Error("checker should not be nil after reset")
	}

	// Restore original for other tests
	diskSpaceCheckerMu.Lock()
	diskSpaceCheckerFn = originalChecker
	diskSpaceCheckerMu.Unlock()
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    uint64
		expected string
	}{
		{0, "0 bytes"},
		{512, "512 bytes"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 10, "10.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 2, "2.0 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := FormatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("FormatBytes(%d) = %q, want %q", tc.bytes, result, tc.expected)
			}
		})
	}
}

func TestCheckDiskSpaceWarning_MemoryDB(t *testing.T) {
	// In-memory database should not produce warnings
	db := &DB{path: ":memory:"}
	warning := db.CheckDiskSpaceWarning()
	if warning != nil {
		t.Error("expected no warning for :memory: database")
	}
}

func TestCheckDiskSpaceWarning_EmptyPath(t *testing.T) {
	// Empty path should not produce warnings
	db := &DB{path: ""}
	warning := db.CheckDiskSpaceWarning()
	if warning != nil {
		t.Error("expected no warning for empty path")
	}
}

func TestDiskSpaceThresholds(t *testing.T) {
	// Verify threshold constants are sensible
	if minFreeDiskBytes >= warnFreeDiskBytes {
		t.Errorf("minFreeDiskBytes (%d) should be less than warnFreeDiskBytes (%d)",
			minFreeDiskBytes, warnFreeDiskBytes)
	}

	// Verify minimum is at least 1MB
	if minFreeDiskBytes < 1024*1024 {
		t.Errorf("minFreeDiskBytes (%d) should be at least 1MB", minFreeDiskBytes)
	}

	// Verify warning threshold is reasonable (at least 5MB)
	if warnFreeDiskBytes < 5*1024*1024 {
		t.Errorf("warnFreeDiskBytes (%d) should be at least 5MB", warnFreeDiskBytes)
	}
}
