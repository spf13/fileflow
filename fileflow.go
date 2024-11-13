package fileflow

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"syscall"
)

type ErrFailedRemovingOriginal struct {
	input error
	file  string
}

func (e ErrFailedRemovingOriginal) Error() string {
	return fmt.Sprintf("failed removing original file %v", e.file)
}

type ErrFailedCopyingFile struct {
	input error
	src   string
	dst   string
}

func (e ErrFailedCopyingFile) Error() string {
	return fmt.Sprintf("failed copying file %v to %v", e.src, e.dst)
}

type ErrFailedMovingFile struct {
	input error
	src   string
	dst   string
}

func (e ErrFailedMovingFile) Error() string {
	return fmt.Sprintf("failed moving file %v to %v", e.src, e.dst)
}

// SafeMoveEfficient moves a file from src to dst, and returns the final destination
// it first tries to rename, if the file is on a different drive, it moves instead
func SafeMoveFileEfficient(src, dst string) (string, error) {
	final, err := SafeRename(src, dst)

	if err != nil {
		terr, ok := err.(*os.LinkError)
		if ok && terr.Err == syscall.EXDEV {
			// if the file is on a different drive, copy it instead
			final, err = SafeMoveFile(src, dst)
		}
	}

	return final, err
}

// SafeMoveFile moves a file from src to dst, and returns the final destination
// It ensures that the dst file is not overwritten unless it is identical to the src file
// It will append -1, -2, -3, etc to the end of the file if it already exists (and isn't identical)
func SafeRename(src, dst string) (string, error) {
	if src == dst {
		return "", fmt.Errorf("source and destination are the same")
	}
	// if file exists, check if it's the same file
	// if it's the same file, remove the old one
	// if it's a different file, name the new one dst-1, dst-2, etc

	// if file exists, check if it's the same file
	if FileExists(dst) {
		if IdenticalFiles(src, dst) { // if it's the same file, remove the old one
			err := os.Remove(src)
			if err != nil {
				return dst, &ErrFailedRemovingOriginal{input: err, file: src}
			}
			return dst, nil // successfully remove the file, return nil
		} else { // if it's a different file, rename the new one
			for x := 1; x < 100; x++ {
				// try to rename the file with -1, -2, -3, etc appended to the end
				newDst := newIncrementedName(dst, x)
				fmt.Println("newDst", newDst)
				if !FileExists(newDst) {
					// if the file doesn't exist, move it there
					newdst, err := SafeRename(src, newDst)
					if err != nil {
						return "", &ErrFailedMovingFile{input: err, src: src, dst: dst}
					}
					return newdst, nil
				}

			}
		}
	}
	{
		err := os.MkdirAll(filepath.Dir(dst), 0755)
		if err != nil {
			return "", fmt.Errorf("failed creating destination path: %s", err)
		}
	}
	// if dst file doesn't exist, move src to it
	err := os.Rename(src, dst)
	if err != nil {
		return "", &ErrFailedMovingFile{input: err, src: src, dst: dst}
	}
	return dst, nil
}

// MoveFile moves a file from src to dst, renaming it if necessary
// It will not overwrite an existing file, but will rename the new file
// to dst-1, dst-2, etc
// Will return the final name of the file written to dst and an error
func SafeMoveFile(src, dst string) (string, error) {
	if src == dst {
		return "", fmt.Errorf("source and destination are the same")
	}
	// if file exists, check if it's the same file
	// if it's the same file, remove the old one
	// if it's a different file, name the new one dst-1, dst-2, etc

	// if file exists, check if it's the same file
	if FileExists(dst) {
		if IdenticalFiles(src, dst) { // if it's the same file, remove the old one
			err := os.Remove(src)
			if err != nil {
				return dst, &ErrFailedRemovingOriginal{input: err, file: src}
			}
			return dst, nil // successfully remove the file, return nil
		} else { // if it's a different file, rename the new one
			for x := 1; x < 100; x++ {
				// try to rename the file with -1, -2, -3, etc appended to the end
				newDst := newIncrementedName(dst, x)
				fmt.Println("newDst", newDst)
				if !FileExists(newDst) {
					// if the file doesn't exist, move it there
					newdst, err := SafeMoveFile(src, newDst)
					if err != nil {
						return "", &ErrFailedMovingFile{input: err, src: src, dst: dst}
					}
					return newdst, nil
				}

			}
		}
	}

	// if file doesn't exist, move it
	err := CopyFileCreatingPaths(src, dst)
	if err != nil {
		fmt.Println("failed copying file", err)
		return "", &ErrFailedCopyingFile{input: err, src: src, dst: dst}
	}

	// Even if the nested remove fails, this will remove it on the parent run.
	err = os.Remove(src)
	if err != nil {
		return dst, &ErrFailedCopyingFile{input: err, src: src, dst: dst}
	}

	return dst, nil
}

// FileExists returns true if the file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

// newIncrementedName returns a new name for a file, appending -1, -2, -3, etc to the end
// replacing any existing -1, -2, -3, etc (only the final -N)
func newIncrementedName(name string, i int) string {
	ext := filepath.Ext(name)
	nameWOExt := name[:len(name)-len(ext)]
	nameWOInc := regexp.MustCompile(`-\d+$`).ReplaceAllString(nameWOExt, "")
	// regular expression to match -1, -2, -3, etc
	// if it matches, remove it
	return nameWOInc + "-" + strconv.Itoa(i) + ext
}

// Compare two files byte by byte and return false as soon as they are different
func IdenticalFiles(file1, file2 string) bool {
	sf, err := os.Open(file1)
	defer sf.Close()

	if err != nil {
		log.Fatal(err)
	}

	df, err := os.Open(file2)
	defer df.Close()

	if err != nil {
		log.Fatal(err)
	}

	sscan := bufio.NewScanner(sf)
	dscan := bufio.NewScanner(df)

	for sscan.Scan() {
		dscan.Scan()
		if !bytes.Equal(sscan.Bytes(), dscan.Bytes()) {
			return false
		}
	}

	return true
}

// MoveFileCreatingPaths moves a file from src to dst, creating the destination path if it doesn't exist
func MoveFileCreatingPaths(src, dst string) error {
	err := CopyFileCreatingPaths(src, dst)
	if err != nil {
		return &ErrFailedCopyingFile{input: err, src: src, dst: dst}
	}
	err = os.Remove(src)
	if err != nil {
		return &ErrFailedRemovingOriginal{input: err, file: src}
	}

	return nil
}

// CopyFileCreatingPaths copies a file from src to dst, creating the destination path if it doesn't exist
func CopyFileCreatingPaths(src, dst string) error {
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return fmt.Errorf("failed creating destination path: %s", err)
	}

	// move the file
	err = CopyFile(src, dst)
	if err != nil {
		return &ErrFailedCopyingFile{input: err, src: src, dst: dst}
	}

	return nil
}

// CopyFile copies a file from src to dst across filesystems
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("couldn't open source file: %s", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return fmt.Errorf("couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("sync error: %s", err)
	}

	si, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat error: %s", err)
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return fmt.Errorf("chmod error: %s", err)
	}

	return nil
}

// MoveFile moves a file from src to dst even across filesystems
func MoveFile(src, dst string) error {
	err := CopyFile(src, dst)

	if err != nil {
		return &ErrFailedCopyingFile{src: src, dst: dst, input: err}
	}

	err = os.Remove(src)

	if err != nil {
		return &ErrFailedRemovingOriginal{input: err, file: src}
	}

	return nil
}
