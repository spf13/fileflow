package fileflow

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"
)

const (
	// DefaultBufferSize is the default buffer size used for file operations
	DefaultBufferSize = 32 * 1024 // 32KB buffer
	// DefaultFileMode is the default permission mode for new files
	DefaultFileMode = 0644
	// DefaultDirMode is the default permission mode for new directories
	DefaultDirMode = 0755
)

var (
	ErrSameFile           = errors.New("source and destination are the same")
	ErrMaxAttemptsReached = errors.New("maximum increment attempts reached")
	ErrLockTimeout        = errors.New("timeout acquiring file lock")
	// MaxIncrementAttempts is the maximum number of attempts to increment a filename
	MaxIncrementAttempts             = 100               // user can override this value
	BufferSize                       = DefaultBufferSize // user can override this value
	FileMode             fs.FileMode = DefaultFileMode   // user can override this value
	DirMode              fs.FileMode = DefaultDirMode    // user can override this value
	// FindAvailableName is the function used to find an available filename
	// The default behavior is to increment the filename
	// User can override this behavior by setting this variable to a custom
	// function or the provided FindAvailableNameTS which instead of
	// incrementing adds a timestamp

	FindAvailableName func(string) (string, error) = FindAvailableNameInc
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

// Move tries to move a file atomically using rename if possible,
// falling back to copy+delete if files are on different filesystems.
func Move(src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	final, err := Rename(src, dst)
	if err != nil {
		var linkErr *os.LinkError
		if errors.As(err, &linkErr) && linkErr.Err == syscall.EXDEV {
			// If the file is on a different drive, copy it instead
			return fileMove(src, dst)
		}
		return "", err
	}

	return final, nil
}

// Rename attempts to rename a file from src to dst, handling naming conflicts.
// It returns the final destination path.
func Rename(src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	if Exists(dst) {
		identical, err := Equal(src, dst)
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
		dst, err = FindAvailableName(dst)
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

// fileMove moves a file from src to dst, handling naming conflicts.
// It ensures that the dst file is not overwritten unless it is identical to the src file.
func fileMove(src, dst string) (string, error) {
	if src == dst {
		return "", ErrSameFile
	}

	if Exists(dst) {
		identical, err := Equal(src, dst)
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
		dst, err = FindAvailableName(dst)
		if err != nil {
			return "", fmt.Errorf("finding available name: %w", err)
		}
	}

	if err := CopyWithPaths(src, dst); err != nil {
		return "", err
	}

	if err := os.Remove(src); err != nil {
		return dst, &ErrFailedRemovingOriginal{err: err, file: src}
	}

	return dst, nil
}

// Exists returns true if the file exists and is accessible
func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

var incrementPattern = regexp.MustCompile(`-\d+$`)

// FindAvailableNameInc returns an available filename by incrementing a counter
func FindAvailableNameInc(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]
	nameWOInc := incrementPattern.ReplaceAllString(nameWOExt, "")

	for i := 1; i <= MaxIncrementAttempts; i++ {
		newName := fmt.Sprintf("%s-%d%s", nameWOInc, i, ext)
		if !Exists(newName) {
			return newName, nil
		}
	}

	return "", ErrMaxAttemptsReached
}

func FindAvailableNameTS(baseName string) (string, error) {
	ext := filepath.Ext(baseName)
	nameWOExt := baseName[:len(baseName)-len(ext)]
	nameWOInc := incrementPattern.ReplaceAllString(nameWOExt, "")

	for i := 1; i <= MaxIncrementAttempts; i++ {
		newName := fmt.Sprintf("%s-%s%s", nameWOInc, time.Now().Format("20060102-150405.000000000"), ext)
		if !Exists(newName) {
			return newName, nil
		}
	}

	return "", ErrMaxAttemptsReached
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

	b1 := make([]byte, BufferSize)
	b2 := make([]byte, BufferSize)

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

// CopyWithPaths copies a file from src to dst, creating any necessary paths.
// Returns the final destination path.
func CopyWithPaths(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), DirMode); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}

	return Copy(src, dst)
}

// Copy performs an efficient copy of a file from src to dst.
// If the destination file exists and is identical, it returns early.
// If the destination exists and is different, it finds an available name.
func Copy(src, dst string) error {
	if src == dst {
		return ErrSameFile
	}

	if Exists(dst) {
		identical, err := Equal(src, dst)
		if err != nil {
			return fmt.Errorf("checking file identity: %w", err)
		}

		if identical {
			return nil // File already exists and is identical
		}

		// Find an available filename
		newDst, err := FindAvailableName(dst)
		if err != nil {
			return fmt.Errorf("finding available name: %w", err)
		}
		dst = newDst
	}

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
	writer := bufio.NewWriterSize(destFile, BufferSize)

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
