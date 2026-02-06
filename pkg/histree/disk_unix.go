//go:build !windows

package histree

import (
	"fmt"
	"syscall"
)

// Disk space thresholds for history recording.
// SQLite with WAL mode needs space for the main DB file, WAL file (.wal), and
// shared memory file (.shm).
const (
	// minFreeDiskBytes is the minimum required - below this, writes are blocked
	minFreeDiskBytes = 1 * 1024 * 1024 // 1MB

	// warnFreeDiskBytes is the warning threshold - below this, a warning is shown
	warnFreeDiskBytes = 10 * 1024 * 1024 // 10MB
)

func platformHasSufficientDiskSpace(dir string) (bool, error) {
	availableBytes, err := platformGetAvailableDiskSpace(dir)
	if err != nil {
		return false, err
	}
	return availableBytes >= minFreeDiskBytes, nil
}

func platformGetAvailableDiskSpace(dir string) (uint64, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0, fmt.Errorf("failed to stat filesystem: %w", err)
	}

	// Calculate available bytes: available blocks * block size
	return uint64(stat.Bavail) * uint64(stat.Bsize), nil
}

func platformShouldWarnDiskSpace(dir string) (bool, uint64, error) {
	availableBytes, err := platformGetAvailableDiskSpace(dir)
	if err != nil {
		return false, 0, err
	}
	return availableBytes < warnFreeDiskBytes && availableBytes >= minFreeDiskBytes, availableBytes, nil
}
