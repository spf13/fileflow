//go:build windows

package fileflow

import (
	"errors"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// errNotSameDevice is Windows' ERROR_NOT_SAME_DEVICE (0x11), returned by
// MoveFile when src and dst are on different volumes.
const errNotSameDevice = syscall.Errno(0x11)

// isCrossDeviceError reports whether err indicates a rename across filesystems.
func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV) || errors.Is(err, errNotSameDevice)
}

// renameNoReplace renames src to dst, failing with an error satisfying
// errors.Is(err, fs.ErrExist) if dst already exists. It never overwrites dst.
//
// It calls MoveFileEx with no flags: without MOVEFILE_REPLACE_EXISTING the
// call fails if dst exists, and without MOVEFILE_COPY_ALLOWED it fails with
// ERROR_NOT_SAME_DEVICE across volumes, which callers detect via
// isCrossDeviceError.
func renameNoReplace(src, dst string) error {
	linkErr := func(err error) error {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}

	from, err := windows.UTF16PtrFromString(src)
	if err != nil {
		return linkErr(err)
	}
	to, err := windows.UTF16PtrFromString(dst)
	if err != nil {
		return linkErr(err)
	}

	// No flags: no replace, no cross-volume copy.
	if err := windows.MoveFileEx(from, to, 0); err != nil {
		return linkErr(err)
	}
	return nil
}
