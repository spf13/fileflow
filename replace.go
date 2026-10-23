package fileflow

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Replace moves the regular file src to exactly dst, atomically replacing an
// existing destination. Unlike Move, Copy, and Rename, it deliberately overwrites
// different content and never selects a suffixed name. It returns dst on success.
//
// Across filesystems it copies into a temporary file in the destination directory
// before publishing, then removes src. If that removal fails, it returns dst and
// an *ErrFailedRemovingOriginal: the replacement has already succeeded.
//
// Source symlinks and non-regular files are rejected with fs.ErrInvalid. A
// destination symlink is replaced, not followed. Paths referring to the same file
// return ErrSameFile. Atomic visibility depends on filesystem rename support;
// Replace does not promise crash durability. Sync and close a staged source before
// calling Replace when its durability matters.
func Replace(src, dst string) (string, error) {
	var f Flow
	return f.Replace(src, dst)
}

// Replace moves src to exactly dst with the semantics of [Replace]. It respects
// DirMode and NoCreateDirs, but does not use FindAvailableName.
func (f *Flow) Replace(src, dst string) (string, error) {
	info, err := os.Lstat(src)
	if err != nil {
		return "", &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}
	if !info.Mode().IsRegular() {
		return "", &ErrFailedMovingFile{Err: fs.ErrInvalid, Src: src, Dst: dst}
	}
	if sameFileReplace(src, dst, info) {
		return "", ErrSameFile
	}
	if !f.NoCreateDirs {
		if err := os.MkdirAll(filepath.Dir(dst), f.dirMode()); err != nil {
			return "", fmt.Errorf("creating destination directory: %w", err)
		}
	}

	if err := renameReplace(src, dst); err == nil {
		return dst, nil
	} else if !isCrossDeviceError(err) {
		return "", &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}

	// Publish only after copying and syncing the full file on the target volume.
	tmp, err := copyToTemp(src, filepath.Dir(dst))
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp)
	if err := renameReplace(tmp, dst); err != nil {
		return "", &ErrFailedMovingFile{Err: err, Src: src, Dst: dst}
	}
	if err := os.Remove(src); err != nil {
		return dst, &ErrFailedRemovingOriginal{Err: err, File: src}
	}
	return dst, nil
}
// sameFileReplace reports whether src and dst refer to the same file without
// following a symlink at dst. srcInfo is the Lstat of src.
func sameFileReplace(src, dst string, srcInfo os.FileInfo) bool {
	if src == dst {
		return true
	}
	dstInfo, err := os.Lstat(dst)
	if err != nil {
		return false
	}
	return os.SameFile(srcInfo, dstInfo)
}
