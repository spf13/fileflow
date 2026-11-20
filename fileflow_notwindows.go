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

// syncDir fsyncs the directory at dir to ensure a newly created or
// renamed entry is durable on storage. Errors EINVAL, EBADF, ENOTSUP,
// and EROFS are ignored (special filesystems that do not support fsync).
func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return nil
	}
	defer f.Close()
	if err := f.Sync(); err != nil && !errors.Is(err, syscall.EINVAL) && !errors.Is(err, syscall.EBADF) && !errors.Is(err, syscall.ENOTSUP) && !errors.Is(err, syscall.EROFS) {
		return err
	}
	return nil
}

// renameNoReplace renames src to dst, failing with an error satisfying
// errors.Is(err, fs.ErrExist) if dst already exists. It never overwrites dst.
//
// os.Rename on Unix overwrites silently. Where the kernel offers an atomic
// no-replace rename (Linux renameat2, macOS renamex_np) it is used. Otherwise
// this uses link+remove: link fails atomically with EEXIST if dst exists. On
// filesystems without hard link support it falls back to a plain rename only
// when allowOverwriteFallback is set, which is best-effort.
func renameNoReplace(src, dst string, allowOverwriteFallback bool) error {
	if done, err := renameExclusive(src, dst); done {
		return err
	}
	return renameLink(src, dst, allowOverwriteFallback)
}

// renameLink is renameNoReplace's fallback when no atomic no-replace rename
// is available: link fails with EEXIST if dst exists, then src is removed.
func renameLink(src, dst string, allowOverwriteFallback bool) error {
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
