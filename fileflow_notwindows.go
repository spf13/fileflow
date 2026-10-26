//go:build !windows

package fileflow

import (
	"errors"
	"io/fs"
	"os"
	"syscall"
)

// isCrossDeviceError reports whether err indicates a rename across filesystems.
func isCrossDeviceError(err error) bool {
	return errors.Is(err, syscall.EXDEV)
}

// renameNoReplace renames src to dst, failing with an error satisfying
// errors.Is(err, fs.ErrExist) if dst already exists. It never overwrites dst.
//
// os.Rename on Unix overwrites silently, so this uses link+remove: link
// fails atomically with EEXIST if dst exists. On filesystems without hard
// link support it falls back to a plain rename, which is best-effort only.
func renameNoReplace(src, dst string, allowOverwriteFallback bool) error {
	err := os.Link(src, dst)
	if err == nil {
		if rmErr := os.Remove(src); rmErr != nil {
			return &ErrFailedRemovingOriginal{Err: rmErr, File: src}
		}
		return nil
	}

	if errors.Is(err, fs.ErrExist) || errors.Is(err, fs.ErrNotExist) || isCrossDeviceError(err) {
		return err
	}

	// The filesystem likely doesn't support hard links (e.g. FAT, some
	// network mounts). Only fall back to plain rename when explicitly
	// allowed; otherwise return the link error to avoid silent overwrite.
	if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.ENOTSUP) ||
		errors.Is(err, syscall.ENOSYS) || errors.Is(err, syscall.EMLINK) {
		if !allowOverwriteFallback {
			return err
		}
		return os.Rename(src, dst)
	}

	return err
}
