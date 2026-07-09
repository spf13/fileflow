//go:build !windows

package fileflow

import (
	"errors"
	"syscall"
)

// isCrossDeviceError reports whether err indicates a rename across filesystems.
func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}
