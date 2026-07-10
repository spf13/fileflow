//go:build windows

package fileflow

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

// errNotSameDevice is Windows' ERROR_NOT_SAME_DEVICE (0x11), returned by
// MoveFile when src and dst are on different volumes.
const errNotSameDevice = syscall.Errno(0x11)

// isCrossDeviceError reports whether err indicates a rename across filesystems.
func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV) || errors.Is(err, errNotSameDevice)
}

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procMoveFileExW = kernel32.NewProc("MoveFileExW")
)

// renameNoReplace renames src to dst, failing with an error satisfying
// errors.Is(err, fs.ErrExist) if dst already exists. It never overwrites dst.
//
// It calls MoveFileExW with no flags: without MOVEFILE_REPLACE_EXISTING the
// call fails if dst exists, and without MOVEFILE_COPY_ALLOWED it fails with
// ERROR_NOT_SAME_DEVICE across volumes, which callers detect via
// isCrossDeviceError.
func renameNoReplace(src, dst string) error {
	linkErr := func(err error) error {
		return &os.LinkError{Op: "rename", Old: src, New: dst, Err: err}
	}

	from, err := syscall.UTF16PtrFromString(src)
	if err != nil {
		return linkErr(err)
	}
	to, err := syscall.UTF16PtrFromString(dst)
	if err != nil {
		return linkErr(err)
	}

	r1, _, e1 := procMoveFileExW.Call(
		uintptr(unsafe.Pointer(from)),
		uintptr(unsafe.Pointer(to)),
		0, // no flags: no replace, no cross-volume copy
	)
	if r1 == 0 {
		if errno, ok := e1.(syscall.Errno); ok && errno != 0 {
			return linkErr(errno)
		}
		return linkErr(syscall.EINVAL)
	}
	return nil
}
