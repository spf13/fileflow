package fileflow

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
)

const (
	// DefaultBufferSize is the default buffer size used for file operations
	DefaultBufferSize = 32 * 1024 // 32KB buffer
	// MaxIncrementAttempts is the maximum number of attempts to increment a filename
	MaxIncrementAttempts = 100
	// DefaultFileMode is the default permission mode for new files
	DefaultFileMode = 0644
	// DefaultDirMode is the default permission mode for new directories
	DefaultDirMode = 0755
)

var (
	ErrSameFile           = errors.New("source and destination are the same")
	ErrMaxAttemptsReached = errors.New("maximum increment attempts reached")
)

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

// ErrFailedCopyingFile occurs when a file copy operation fails
type ErrFailedCopyingFile struct {
	err error
	src string
	dst string
}

func (e *ErrFailedCopyingFile) Error() string {
	return fmt.Sprintf("failed copying file %v to %v: %v", e.src, e.dst, e.err)
}

func (e *ErrFailedCopyingFile) Unwrap() error {
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

// SafeMoveFileEfficient moves a file from src to dst efficiently, and returns the final destination.
// It first attempts to rename the file, falling back to copy+delete if the files are on different filesystems.
func SafeMoveFileEfficient(ctx context.Context, src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	final, err := SafeRename(ctx, src, dst)
	if err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && linkErr.Err == syscall.EXDEV {
			// If the file is on a different drive, copy it instead
			return SafeMoveFile(ctx, src, dst)
		}
		return "", err
	}

	return final, nil
}

// SafeRename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path.
func SafeRename(ctx context.Context, src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	if FileExists(dst) {
		identical, err := IdenticalFiles(src, dst)
		if err != nil {
			return "", fmt.Errorf("checking file identity: %w", err)
		}

		if identical {
			if err := os.Remove(src); err != nil {
				return dst, &ErrFailedRemovingOriginal{err: err, file: src}
			}
			return dst, nil
		}

		// Find an available filename
		dst, err = findAvailableName(dst)
		if err != nil {
			return "", fmt.Errorf("finding available name: %w", err)
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), DefaultDirMode); err != nil {
		return "", fmt.Errorf("creating destination directory: %w", err)
	}

	if err := os.Rename(src, dst); err != nil {
		return "", &ErrFailedMovingFile{err: err, src: src, dst: dst}
	}

	return dst, nil
}

// SafeMoveFile moves a file from src to dst, handling naming conflicts.
// It ensures that the dst file is not overwritten unless it is identical to the src file.
func SafeMoveFile(ctx context.Context, src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	if FileExists(dst) {
		identical, err := IdenticalFiles(src, dst)
		if err != nil {
			return "", fmt.Errorf("checking file identity: %w", err)
		}

		if identical {
			if err := os.Remove(src); err != nil {
				return dst, &ErrFailedRemovingOriginal{err: err, file: src}
			}
			return dst, nil
		}

		// Find an available filename
		dst, err = findAvailableName(dst)
		if err != nil {
			return "", fmt.Errorf("finding available name: %w", err)
		}
	}

	if err := CopyFileCreatingPaths(ctx, src, dst); err != nil {
		return "", err
	}

	if err := os.Remove(src); err != nil {
		return dst, &ErrFailedRemovingOriginal{err: err, file: src}
	}

	return dst, nil
}

// FileExists returns true if the file exists and is accessible
func FileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

var incrementPattern = regexp.MustCompile(`-\d+$`)

// findAvailableName returns an available filename by incrementing a counter
func findAvailableName(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]
	nameWOInc := incrementPattern.ReplaceAllString(nameWOExt, "")

	for i := 1; i <= MaxIncrementAttempts; i++ {
		newName := fmt.Sprintf("%s-%d%s", nameWOInc, i, ext)
		if !FileExists(newName) {
			return newName, nil
		}
	}

	return "", ErrMaxAttemptsReached
}

// IdenticalFiles compares two files and returns true if they have identical content
func IdenticalFiles(file1, file2 string) (bool, error) {
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

	const chunkSize = DefaultBufferSize
	b1 := make([]byte, chunkSize)
	b2 := make([]byte, chunkSize)

	for {
		n1, err1 := f1.Read(b1)
		n2, err2 := f2.Read(b2)

		if n1 != n2 || !bytes.Equal(b1[:n1], b2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if err1 != nil && err1 != io.EOF {
			return false, fmt.Errorf("reading first file: %w", err1)
		}
		if err2 != nil && err2 != io.EOF {
			return false, fmt.Errorf("reading second file: %w", err2)
		}
	}
}

// CopyFileCreatingPaths copies a file from src to dst, creating the destination path if needed
func CopyFileCreatingPaths(ctx context.Context, src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), DefaultDirMode); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	return CopyFile(ctx, src, dst)
}

// CopyFile performs an efficient copy of a file from src to dst
func CopyFile(ctx context.Context, src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("getting source file info: %w", err)
	}

	// Create destination file with same permissions
	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}

	// Use buffered writer for better performance
	writer := bufio.NewWriterSize(destFile, DefaultBufferSize)

	// Copy the file
	if _, err := io.Copy(writer, sourceFile); err != nil {
		destFile.Close()
		return fmt.Errorf("copying file content: %w", err)
	}

	// Ensure all buffered data is written
	if err := writer.Flush(); err != nil {
		destFile.Close()
		return fmt.Errorf("flushing writer: %w", err)
	}

	// Ensure file is properly written to disk
	if err := destFile.Sync(); err != nil {
		destFile.Close()
		return fmt.Errorf("syncing file: %w", err)
	}

	if err := destFile.Close(); err != nil {
		return fmt.Errorf("closing destination file: %w", err)
	}

	return nil
}
