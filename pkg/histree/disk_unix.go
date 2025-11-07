//go:build !windows

package histree

import (
	"fmt"
	"syscall"
)

func platformHasSufficientDiskSpace(dir string) (bool, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return false, fmt.Errorf("failed to stat filesystem: %w", err)
	}

	return stat.Bavail > 0, nil
}
