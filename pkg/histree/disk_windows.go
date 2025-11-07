//go:build windows

package histree

import (
	"fmt"
	"syscall"
)

func platformHasSufficientDiskSpace(dir string) (bool, error) {
	pathPtr, err := syscall.UTF16PtrFromString(dir)
	if err != nil {
		return false, fmt.Errorf("failed to encode path for disk query: %w", err)
	}

	var freeBytesAvailable uint64
	if err := syscall.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, nil, nil); err != nil {
		return false, fmt.Errorf("failed to query free disk space: %w", err)
	}

	return freeBytesAvailable > 0, nil
}
