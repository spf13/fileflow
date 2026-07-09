//go:build windows

package fileflow

import (
	"errors"
	"syscall"
)

// errNotSameDevice is Windows' ERROR_NOT_SAME_DEVICE (0x11), returned by
// MoveFile when src and dst are on different volumes.
const errNotSameDevice = syscall.Errno(0x11)

// isCrossDeviceError reports whether err indicates a rename across filesystems.
func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV) || errors.Is(err, errNotSameDevice)
}
