/*
Copyright © 2024 The FileFlow Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package fileflow

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

const (
	// DefaultDirMode is the default permission mode for new directories.
	DefaultDirMode = 0755

	// maxIncrementAttempts is the maximum number of candidate names tried
	// when resolving a naming conflict by incrementing a counter.
	maxIncrementAttempts = 100
)

var (
	ErrSameFile           = errors.New("source and destination are the same")
	ErrMaxAttemptsReached = errors.New("maximum increment attempts reached")
)

// contSuffixPattern matches a trailing "-N" counter (e.g. "-12") at the end
// of a filename (extension already removed). Capture group 1 holds the digits.
var contSuffixPattern = regexp.MustCompile(`-(\d+)$`)

// Flow performs file operations with configurable behavior.
// The zero value uses sensible defaults and is ready to use.
// Configure a Flow before first use; do not mutate it concurrently
// with calls to its methods.
type Flow struct {
	// FindAvailableName resolves naming conflicts when a destination
	// already exists with different content.
	// Nil means FindAvailableNameInc.
	FindAvailableName func(string) (string, error)

	// DirMode is the permission mode used when creating destination
	// directories. Zero means 0755.
	DirMode fs.FileMode
}

// findAvailableName resolves the configured naming strategy, defaulting to
// FindAvailableNameInc when none is set.
func (f *Flow) findAvailableName(baseName string) (string, error) {
	if f.FindAvailableName != nil {
		return f.FindAvailableName(baseName)
	}
	return FindAvailableNameInc(baseName)
}

// dirMode resolves the directory permission mode, defaulting to DefaultDirMode
// when unset.
func (f *Flow) dirMode() fs.FileMode {
	if f.DirMode != 0 {
		return f.DirMode
	}
	return DefaultDirMode
}

// ErrFailedRemovingOriginal occurs when the original file cannot be removed
type ErrFailedRemovingOriginal struct {
	err  error
	file string
}

func (e *ErrFailedRemovingOriginal) Error() string {
	return fmt.Sprintf("failed removing original file %v: %v", e.file, e.err)
}

func (e *ErrFailedRemovingOriginal) Unwrap() error {
	return e.err
}

// ErrFailedMovingFile occurs when a file move operation fails
type ErrFailedMovingFile struct {
	err error
	src string
	dst string
}

func (e *ErrFailedMovingFile) Error() string {
	return fmt.Sprintf("failed moving file %v to %v: %v", e.src, e.dst, e.err)
}

func (e *ErrFailedMovingFile) Unwrap() error {
	return e.err
}

// sameFile reports whether src and dst refer to the same underlying file,
// even when the paths are spelled differently (e.g. "./a" vs "a", symlinks,
// or case differences on case-insensitive filesystems).
func sameFile(src, dst string) bool {
	if src == dst {
		return true
	}
	srcInfo, err := os.Stat(src)
	if err != nil {
		return false
	}
	dstInfo, err := os.Stat(dst)
	if err != nil {
		return false
	}
	return os.SameFile(srcInfo, dstInfo)
}

// Move tries to move a file atomically using rename if possible,
// falling back to copy+delete if files are on different filesystems.
func Move(src, dst string) (string, error) {
	var f Flow
	return f.Move(src, dst)
}

// Move tries to move a file atomically using rename if possible,
// falling back to copy+delete if files are on different filesystems.
func (f *Flow) Move(src, dst string) (string, error) {
	if sameFile(src, dst) {
		return "", ErrSameFile
	}

	final, err := f.Rename(src, dst)
	if err != nil {
		if isCrossDeviceError(err) {
			// If the file is on a different drive, copy it instead
			return f.fileMove(src, dst)
		}
		return "", err
	}

	return final, nil
}

// resolveDst decides where an operation on src may write. It returns the
// destination path to use and whether dst already holds content identical
// to src (in which case no write is needed).
func (f *Flow) resolveDst(src, dst string) (path string, identical bool, err error) {
	if sameFile(src, dst) {
		return "", false, ErrSameFile
	}
	if !Exists(dst) {
		return dst, false, nil
	}
	same, err := Equal(src, dst)
	if err != nil {
		return "", false, fmt.Errorf("checking file identity: %w", err)
	}
	if same {
		return dst, true, nil
	}
	path, err = f.findAvailableName(dst)
	if err != nil {
		return "", false, fmt.Errorf("finding available name: %w", err)
	}
	return path, false, nil
}

// Rename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path.
func Rename(src, dst string) (string, error) {
	var f Flow
	return f.Rename(src, dst)
}

// Rename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path.
func (f *Flow) Rename(src, dst string) (string, error) {
	path, identical, err := f.resolveDst(src, dst)
	if err != nil {
		return "", err
	}

	if identical {
		if err := os.Remove(src); err != nil {
			return dst, &ErrFailedRemovingOriginal{err: err, file: src}
		}
		return dst, nil
	}

	if err := os.MkdirAll(filepath.Dir(path), f.dirMode()); err != nil {
		return "", fmt.Errorf("creating destination directory: %w", err)
	}

	if err := os.Rename(src, path); err != nil {
		return "", &ErrFailedMovingFile{err: err, src: src, dst: path}
	}

	return path, nil
}

// fileMove moves a file across filesystems by copy + remove.
func (f *Flow) fileMove(src, dst string) (string, error) {
	final, err := f.CopyWithPaths(src, dst)
	if err != nil {
		return "", err
	}
	if err := os.Remove(src); err != nil {
		return final, &ErrFailedRemovingOriginal{err: err, file: src}
	}
	return final, nil
}

// Exists returns true if the file exists and is accessible
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// FindAvailableNameInc returns an available filename by incrementing a counter
func FindAvailableNameInc(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]

	for i := 1; i <= maxIncrementAttempts; i++ {
		newName := fmt.Sprintf("%s-%d%s", nameWOExt, i, ext)
		if !Exists(newName) {
			return newName, nil
		}
	}

	return "", ErrMaxAttemptsReached
}

// FindAvailableNameCont returns an available filename by treating a
// trailing -N suffix as a counter and continuing the sequence from N+1.
// A name without a trailing -N behaves like FindAvailableNameInc.
// Do not use this strategy when filenames end in meaningful digits
// (e.g. years): "report-2024.txt" would continue as "report-2025.txt".
func FindAvailableNameCont(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]

	base := nameWOExt
	start := 1
	if m := contSuffixPattern.FindStringSubmatch(nameWOExt); m != nil {
		if n, err := strconv.Atoi(m[1]); err == nil {
			base = nameWOExt[:len(nameWOExt)-len(m[0])]
			start = n + 1
		}
	}

	for attempt := 0; attempt < maxIncrementAttempts; attempt++ {
		candidate := fmt.Sprintf("%s-%d%s", base, start+attempt, ext)
		if !Exists(candidate) {
			return candidate, nil
		}
	}

	return "", ErrMaxAttemptsReached
}

// FindAvailableNameAuto chooses between the Inc and Cont strategies based
// on the trailing suffix. A trailing -N counter of one or two digits
// (e.g. "file-1", "file-42") is treated as a counter and continued from
// N+1, like FindAvailableNameCont. Three or more trailing digits
// (e.g. "report-2024", "img-001") are assumed to be meaningful and are
// preserved, appending a new counter like FindAvailableNameInc.
func FindAvailableNameAuto(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]

	if m := contSuffixPattern.FindStringSubmatch(nameWOExt); m != nil && len(m[1]) <= 2 {
		return FindAvailableNameCont(baseName)
	}
	return FindAvailableNameInc(baseName)
}

// FindAvailableNameTS returns an available filename by appending a timestamp.
func FindAvailableNameTS(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]
	candidate := fmt.Sprintf("%s-%s%s", nameWOExt, time.Now().Format("20060102-150405.000000000"), ext)

	if !Exists(candidate) {
		return candidate, nil
	}

	return FindAvailableNameInc(candidate)
}

// Equal compares two files and returns true if they have identical content
func Equal(file1, file2 string) (bool, error) {
	f1Info, err := os.Stat(file1)
	if err != nil {
		return false, fmt.Errorf("stat file1: %w", err)
	}
	f2Info, err := os.Stat(file2)
	if err != nil {
		return false, fmt.Errorf("stat file2: %w", err)
	}

	// Quick check: if sizes differ, files are not identical
	if f1Info.Size() != f2Info.Size() {
		return false, nil
	}

	f1, err := os.Open(file1)
	if err != nil {
		return false, fmt.Errorf("opening first file: %w", err)
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, fmt.Errorf("opening second file: %w", err)
	}
	defer f2.Close()

	return readersEqual(f1, f2)
}

func readersEqual(r1, r2 io.Reader) (bool, error) {
	b1 := bufio.NewReader(r1)
	b2 := bufio.NewReader(r2)

	for {
		byte1, err1 := b1.ReadByte()
		byte2, err2 := b2.ReadByte()

		if err1 != nil && err1 != io.EOF {
			return false, fmt.Errorf("reading first file: %w", err1)
		}
		if err2 != nil && err2 != io.EOF {
			return false, fmt.Errorf("reading second file: %w", err2)
		}
		if err1 == io.EOF || err2 == io.EOF {
			return err1 == io.EOF && err2 == io.EOF, nil
		}
		if byte1 != byte2 {
			return false, nil
		}
	}
}

// CopyWithPaths copies a file from src to dst, creating any necessary paths.
// Returns the final destination path, which may differ from dst if a naming
// conflict was resolved.
func CopyWithPaths(src, dst string) (string, error) {
	var f Flow
	return f.CopyWithPaths(src, dst)
}

// CopyWithPaths copies a file from src to dst, creating any necessary paths.
// Returns the final destination path, which may differ from dst if a naming
// conflict was resolved.
func (f *Flow) CopyWithPaths(src, dst string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), f.dirMode()); err != nil {
		return "", fmt.Errorf("creating destination directory: %w", err)
	}

	return f.Copy(src, dst)
}

// Copy performs an efficient copy of a file from src to dst.
// If the destination file exists and is identical, it returns early.
// If the destination exists and is different, it finds an available name.
// It returns the final destination path written, which may differ from dst
// when a naming conflict was resolved. On error it returns an empty path.
func Copy(src, dst string) (string, error) {
	var f Flow
	return f.Copy(src, dst)
}

// Copy performs an efficient copy of a file from src to dst.
// If the destination file exists and is identical, it returns early.
// If the destination exists and is different, it finds an available name.
// It returns the final destination path written, which may differ from dst
// when a naming conflict was resolved. On error it returns an empty path.
func (f *Flow) Copy(src, dst string) (string, error) {
	path, identical, err := f.resolveDst(src, dst)
	if err != nil {
		return "", err
	}
	if identical {
		return dst, nil // File already exists and is identical
	}
	dst = path

	sourceFile, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return "", fmt.Errorf("getting source file info: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".fileflow-*")
	if err != nil {
		return "", fmt.Errorf("creating temporary destination file: %w", err)
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()

	if _, err := io.Copy(tmp, sourceFile); err != nil {
		return "", fmt.Errorf("copying file content: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return "", fmt.Errorf("syncing file: %w", err)
	}

	if err := os.Chmod(tmpName, sourceInfo.Mode()); err != nil {
		return "", fmt.Errorf("setting destination file mode: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return "", fmt.Errorf("closing temporary destination file: %w", err)
	}

	if err := os.Rename(tmpName, dst); err != nil {
		return "", fmt.Errorf("renaming temporary destination file: %w", err)
	}

	success = true
	return dst, nil
}
