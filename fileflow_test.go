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
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"syscall"
	"testing"
)

type shortReader struct {
	r   io.Reader
	max int
}

func (r shortReader) Read(p []byte) (int, error) {
	if len(p) > r.max {
		p = p[:r.max]
	}
	return r.r.Read(p)
}

func TestIsCrossDeviceError(t *testing.T) {
	exdevLink := &os.LinkError{Op: "rename", Err: syscall.EXDEV}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"exdev link error", exdevLink, true},
		{"wrapped in ErrFailedMovingFile", &ErrFailedMovingFile{err: exdevLink, src: "a", dst: "b"}, true},
		{"wrapped with fmt.Errorf", fmt.Errorf("moving: %w", exdevLink), true},
		{"other error", errors.New("other"), false},
		{"not exist", os.ErrNotExist, false},
	}

	for _, tt := range tests {
		if got := isCrossDeviceError(tt.err); got != tt.want {
			t.Errorf("isCrossDeviceError(%v) = %v; want %v", tt.err, got, tt.want)
		}
	}

	if runtime.GOOS == "windows" {
		notSameDevice := &os.LinkError{Op: "rename", Err: syscall.Errno(0x11)}
		if got := isCrossDeviceError(notSameDevice); !got {
			t.Errorf("isCrossDeviceError(ERROR_NOT_SAME_DEVICE) = %v; want true", got)
		}
	}
}

func TestFileMove(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	src := filepath.Join(srcDir, "source.txt")
	dst := filepath.Join(dstDir, "dest.txt")
	content := []byte("Hello World")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := fileMove(src, dst)
	if err != nil {
		t.Fatalf("fileMove() error: %v", err)
	}

	if final != dst {
		t.Errorf("fileMove() = %v; want %v", final, dst)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file still exists after fileMove")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading destination file error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("destination content = %q; want %q", got, content)
	}
}

func TestFileMoveIdenticalRemovesSrc(t *testing.T) {
	tempDir := t.TempDir()

	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, content, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := fileMove(src, dst)
	if err != nil {
		t.Fatalf("fileMove() error: %v", err)
	}

	if final != dst {
		t.Errorf("fileMove() = %v; want %v", final, dst)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source file still exists after fileMove of identical file")
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading destination file error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("destination content = %q; want %q", got, content)
	}

	if Exists(filepath.Join(tempDir, "dest-1.txt")) {
		t.Errorf("fileMove() created a -1 file for an identical destination")
	}
}

func TestFileMoveConflict(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	srcContent := []byte("source content")
	dstContent := []byte("destination content")

	if err := os.WriteFile(src, srcContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, dstContent, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := fileMove(src, dst)
	if err != nil {
		t.Fatalf("fileMove() error: %v", err)
	}
	want := filepath.Join(tempDir, "dest-1.txt")
	if final != want {
		t.Errorf("fileMove() = %v; want %v", final, want)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists after fileMove(): %v", err)
	}

	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("reading fileMove() destination: %v", err)
	}
	if !bytes.Equal(got, srcContent) {
		t.Errorf("fileMove() destination content = %q; want %q", got, srcContent)
	}
	original, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading original destination after fileMove(): %v", err)
	}
	if !bytes.Equal(original, dstContent) {
		t.Errorf("original destination content after fileMove() = %q; want %q", original, dstContent)
	}
}

func TestMoveSameFileError(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "a.txt")
	content := []byte("do not lose me")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Move(path, path)
	if !errors.Is(err, ErrSameFile) {
		t.Errorf("Move() error = %v; want %v", err, ErrSameFile)
	}
	if final != "" {
		t.Errorf("Move() = %q; want empty string", final)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file missing after Move() onto itself: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("file content after Move() onto itself = %q; want %q", got, content)
	}
}

func TestRenameSameFileError(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "a.txt")
	content := []byte("do not lose me")

	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Rename(path, path)
	if !errors.Is(err, ErrSameFile) {
		t.Errorf("Rename() error = %v; want %v", err, ErrSameFile)
	}
	if final != "" {
		t.Errorf("Rename() = %q; want empty string", final)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file missing after Rename() onto itself: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("file content after Rename() onto itself = %q; want %q", got, content)
	}
}

func TestMoveSameFileDifferentPath(t *testing.T) {
	// Moving a file onto itself via a differently-spelled path must not
	// destroy the file. E.g. "dir/./a.txt" -> "dir/a.txt".
	tempDir := t.TempDir()

	dst := filepath.Join(tempDir, "a.txt")
	// Same file, different spelling. Built by hand so filepath.Join can't clean it.
	src := tempDir + string(os.PathSeparator) + "." + string(os.PathSeparator) + "a.txt"
	if src == dst {
		t.Fatal("test bug: paths are spelled identically")
	}

	content := []byte("do not lose me")
	if err := os.WriteFile(dst, content, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := Move(src, dst)
	if !errors.Is(err, ErrSameFile) {
		t.Errorf("Move() error = %v; want %v", err, ErrSameFile)
	}
	if final != "" {
		t.Errorf("Move() = %q; want empty string", final)
	}

	// The file must still exist with its content intact.
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("file destroyed by Move onto itself: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("file content = %q; want %q", got, content)
	}
}

func TestRenameSameFileDifferentPath(t *testing.T) {
	tempDir := t.TempDir()

	dst := filepath.Join(tempDir, "a.txt")
	src := tempDir + string(os.PathSeparator) + "." + string(os.PathSeparator) + "a.txt"
	if src == dst {
		t.Fatal("test bug: paths are spelled identically")
	}

	content := []byte("do not lose me")
	if err := os.WriteFile(dst, content, 0644); err != nil {
		t.Fatal(err)
	}

	if _, err := Rename(src, dst); err != nil && !errors.Is(err, ErrSameFile) {
		t.Fatalf("Rename() unexpected error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("file destroyed by Rename onto itself: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("file content = %q; want %q", got, content)
	}
}

func TestIdenticalFiles(t *testing.T) {
	// Set up test files
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "test_file1.txt")
	file2 := filepath.Join(tempDir, "test_file2.txt")
	file3 := filepath.Join(tempDir, "test_file3.txt")

	content := []byte("Hello World")
	diffContent := []byte("Hello World!")

	// Create test files
	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(file3, diffContent, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		fileA, fileB string
		expected     bool
	}{
		{file1, file2, true},
		{file1, file3, false},
		{file2, file3, false},
	}

	for _, tt := range tests {
		result, err := Equal(tt.fileA, tt.fileB)
		if err != nil {
			t.Errorf("IdenticalFiles(%q, %q) error: %v", tt.fileA, tt.fileB, err)
			continue
		}
		if result != tt.expected {
			t.Errorf("IdenticalFiles(%q, %q) = %v; want %v", tt.fileA, tt.fileB, result, tt.expected)
		}
	}
}

func TestReadersEqualShortReads(t *testing.T) {
	data := bytes.Repeat([]byte("fileflow short read test data\n"), 256)
	different := append([]byte(nil), data...)
	different[len(different)-1] ^= 0xff

	tests := []struct {
		name string
		r1   io.Reader
		r2   io.Reader
		want bool
	}{
		{
			name: "identical with one-byte reads",
			r1:   shortReader{r: bytes.NewReader(data), max: 1},
			r2:   bytes.NewReader(data),
			want: true,
		},
		{
			name: "identical with seven-byte reads",
			r1:   bytes.NewReader(data),
			r2:   shortReader{r: bytes.NewReader(data), max: 7},
			want: true,
		},
		{
			name: "different final byte",
			r1:   shortReader{r: bytes.NewReader(data), max: 7},
			r2:   bytes.NewReader(different),
			want: false,
		},
		{
			name: "one stream is a prefix",
			r1:   shortReader{r: bytes.NewReader(data[:len(data)-1]), max: 1},
			r2:   bytes.NewReader(data),
			want: false,
		},
		{
			name: "empty streams",
			r1:   bytes.NewReader(nil),
			r2:   bytes.NewReader(nil),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := readersEqual(tt.r1, tt.r2)
			if err != nil {
				t.Fatalf("readersEqual() error: %v", err)
			}
			if got != tt.want {
				t.Errorf("readersEqual() = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestEqualLargeFiles(t *testing.T) {
	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "file1.bin")
	file2 := filepath.Join(tempDir, "file2.bin")

	data := make([]byte, 100*1024)
	for i := range data {
		data[i] = byte(i % 251)
	}

	if err := os.WriteFile(file1, data, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, data, 0644); err != nil {
		t.Fatal(err)
	}

	equal, err := Equal(file1, file2)
	if err != nil {
		t.Fatalf("Equal() error: %v", err)
	}
	if !equal {
		t.Error("Equal() = false for identical large files; want true")
	}

	different := append([]byte(nil), data...)
	different[len(different)-1] ^= 0xff
	if err := os.WriteFile(file2, different, 0644); err != nil {
		t.Fatal(err)
	}

	equal, err = Equal(file1, file2)
	if err != nil {
		t.Fatalf("Equal() error: %v", err)
	}
	if equal {
		t.Error("Equal() = true for large files differing in final byte; want false")
	}
}

func TestCopy(t *testing.T) {
	// Create a temporary directory for test files
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := Copy(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if final != dstPath {
		t.Errorf("Copy() = %v; want %v", final, dstPath)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Reading destination file error: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination file content = %s; want %s", dstContent, content)
	}
}

func TestCopyPreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skipf("mode bits are not faithfully represented on %s", runtime.GOOS)
	}

	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("Hello World"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Copy(srcPath, dstPath); err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatalf("stat destination: %v", err)
	}
	if got, want := info.Mode().Perm(), os.FileMode(0600); got != want {
		t.Errorf("destination mode = %v; want %v", got, want)
	}
}

func TestCopyLeavesNoTempOnSuccess(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "source.txt")
	dstPath := filepath.Join(dstDir, "dest.txt")

	if err := os.WriteFile(srcPath, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := Copy(srcPath, dstPath); err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	temps, err := filepath.Glob(filepath.Join(dstDir, ".fileflow-*"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(temps) != 0 {
		t.Errorf("temporary files left after Copy(): %v", temps)
	}

	entries, err := os.ReadDir(dstDir)
	if err != nil {
		t.Fatalf("read destination directory: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != filepath.Base(dstPath) {
		t.Errorf("destination directory entries = %v; want only %q", entries, filepath.Base(dstPath))
	}
}

func TestCopyConflictReturnsFinalPath(t *testing.T) {
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	srcContent := []byte("Hello World")
	dstContent := []byte("Different content")

	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, dstContent, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := Copy(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	wantPath := filepath.Join(tempDir, "dest-1.txt")
	if final != wantPath {
		t.Errorf("Copy() = %v; want %v", final, wantPath)
	}

	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("reading final destination file error: %v", err)
	}
	if !bytes.Equal(got, srcContent) {
		t.Errorf("final file content = %q; want %q", got, srcContent)
	}

	// Original destination must be untouched.
	orig, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("reading original destination file error: %v", err)
	}
	if !bytes.Equal(orig, dstContent) {
		t.Errorf("original destination content = %q; want %q", orig, dstContent)
	}
}

func TestCopyIdenticalReturnsDst(t *testing.T) {
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dstPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	final, err := Copy(srcPath, dstPath)
	if err != nil {
		t.Fatalf("Copy() error: %v", err)
	}

	if final != dstPath {
		t.Errorf("Copy() = %v; want %v", final, dstPath)
	}

	if Exists(filepath.Join(tempDir, "dest-1.txt")) {
		t.Errorf("Copy() created a -1 file for an identical destination")
	}
}

func TestCopySamePath(t *testing.T) {
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	if err := os.WriteFile(srcPath, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	final, err := Copy(srcPath, srcPath)
	if !errors.Is(err, ErrSameFile) {
		t.Errorf("Copy() error = %v; want %v", err, ErrSameFile)
	}
	if final != "" {
		t.Errorf("Copy() = %q; want empty string", final)
	}
}

func TestExists(t *testing.T) {
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistingFile := filepath.Join(tempDir, "does_not_exist.txt")

	if err := os.WriteFile(existingFile, []byte("Hello"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		file     string
		expected bool
	}{
		{existingFile, true},
		{nonExistingFile, false},
	}

	for _, tt := range tests {
		result := Exists(tt.file)
		if result != tt.expected {
			t.Errorf("FileExists(%q) = %v; want %v", tt.file, result, tt.expected)
		}
	}
}

func TestExistsOnDirectory(t *testing.T) {
	tempDir := t.TempDir()

	if Exists(tempDir) {
		t.Errorf("Exists(%q) = true for a directory; want false", tempDir)
	}
}

func TestMove(t *testing.T) {
	tempDir := t.TempDir()

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	finalDst, err := Move(srcPath, dstPath)
	if err != nil {
		t.Fatalf("SafeMoveFileEfficient() error: %v", err)
	}

	if finalDst != dstPath {
		t.Errorf("SafeMoveFileEfficient() = %v; want %v", finalDst, dstPath)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Reading destination file error: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination file content = %s; want %s", dstContent, content)
	}
}

func TestMoveConflict(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	srcContent := []byte("source content")
	dstContent := []byte("destination content")

	if err := os.WriteFile(src, srcContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, dstContent, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Move(src, dst)
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}

	want := filepath.Join(tempDir, "dest-1.txt")
	if final != want {
		t.Errorf("Move() = %v; want %v", final, want)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists after Move(): %v", err)
	}
	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("reading Move() destination: %v", err)
	}
	if !bytes.Equal(got, srcContent) {
		t.Errorf("Move() destination content = %q; want %q", got, srcContent)
	}
	original, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading original destination after Move(): %v", err)
	}
	if !bytes.Equal(original, dstContent) {
		t.Errorf("original destination content after Move() = %q; want %q", original, dstContent)
	}
}

func TestMoveIdenticalDst(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	content := []byte("identical content")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, content, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Move(src, dst)
	if err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	if final != dst {
		t.Errorf("Move() = %v; want %v", final, dst)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source still exists after Move() to identical destination: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading destination after Move(): %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("destination content after Move() = %q; want %q", got, content)
	}
}

func TestRenameCreatesDirs(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "nested", "nonexistent", "dir", "dest.txt")
	content := []byte("rename me")

	if err := os.WriteFile(src, content, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Rename(src, dst)
	if err != nil {
		t.Fatalf("Rename() error: %v", err)
	}
	if final != dst {
		t.Errorf("Rename() = %v; want %v", final, dst)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading destination after Rename(): %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("destination content after Rename() = %q; want %q", got, content)
	}
}

func TestRenameConflict(t *testing.T) {
	tempDir := t.TempDir()
	src := filepath.Join(tempDir, "source.txt")
	dst := filepath.Join(tempDir, "dest.txt")
	srcContent := []byte("source content")
	dstContent := []byte("destination content")

	if err := os.WriteFile(src, srcContent, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, dstContent, 0644); err != nil {
		t.Fatal(err)
	}
	final, err := Rename(src, dst)
	if err != nil {
		t.Fatalf("Rename() error: %v", err)
	}

	want := filepath.Join(tempDir, "dest-1.txt")
	if final != want {
		t.Errorf("Rename() = %v; want %v", final, want)
	}
	got, err := os.ReadFile(final)
	if err != nil {
		t.Fatalf("reading Rename() destination: %v", err)
	}
	if !bytes.Equal(got, srcContent) {
		t.Errorf("Rename() destination content = %q; want %q", got, srcContent)
	}
	original, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("reading original destination after Rename(): %v", err)
	}
	if !bytes.Equal(original, dstContent) {
		t.Errorf("original destination content after Rename() = %q; want %q", original, dstContent)
	}
}

func TestCustomFindAvailableName(t *testing.T) {
	tempDir := t.TempDir()

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create a file to ensure the name needs to be incremented
	if err := os.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	// Define a custom function for finding available names
	customFindAvailableName := func(baseName string) (string, error) {
		return baseName + "_custom", nil
	}

	// Set the custom function
	FindAvailableName = customFindAvailableName
	t.Cleanup(func() { FindAvailableName = FindAvailableNameInc })

	newName, err := FindAvailableName(baseName)
	if err != nil {
		t.Fatalf("CustomFindAvailableName() error: %v", err)
	}

	expectedName := baseName + "_custom"
	if newName != expectedName {
		t.Errorf("CustomFindAvailableName() = %v; want %v", newName, expectedName)
	}
}

func TestFindAvailableNameInc(t *testing.T) {
	tempDir := t.TempDir()

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create a file to ensure the name needs to be incremented
	if err := os.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName, err := FindAvailableNameInc(baseName)
	if err != nil {
		t.Fatalf("FindAvailableNameInc() error: %v", err)
	}

	expectedName := baseName[:len(baseName)-4] + "-1.txt"
	if newName != expectedName {
		t.Errorf("FindAvailableNameInc() = %v; want %v", newName, expectedName)
	}

	// Create -1 file and test -2
	if err := os.WriteFile(newName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName2, err := FindAvailableNameInc(baseName)
	if err != nil {
		t.Fatalf("FindAvailableNameInc() second call error: %v", err)
	}

	expectedName2 := baseName[:len(baseName)-4] + "-2.txt"
	if newName2 != expectedName2 {
		t.Errorf("FindAvailableNameInc() second call = %v; want %v", newName2, expectedName2)
	}

	// Test file without extension
	baseNameNoExt := filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(baseNameNoExt, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName3, err := FindAvailableNameInc(baseNameNoExt)
	if err != nil {
		t.Fatalf("FindAvailableNameInc() no extension error: %v", err)
	}

	expectedName3 := baseNameNoExt + "-1"
	if newName3 != expectedName3 {
		t.Errorf("FindAvailableNameInc() no extension = %v; want %v", newName3, expectedName3)
	}

	// A user-supplied name with trailing digits must be preserved, not
	// stripped. Collision on "report-2024.txt" must yield "report-2024-1.txt".
	reportName := filepath.Join(tempDir, "report-2024.txt")
	if err := os.WriteFile(reportName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newReport, err := FindAvailableNameInc(reportName)
	if err != nil {
		t.Fatalf("FindAvailableNameInc() report error: %v", err)
	}

	expectedReport := filepath.Join(tempDir, "report-2024-1.txt")
	if newReport != expectedReport {
		t.Errorf("FindAvailableNameInc() = %v; want %v", newReport, expectedReport)
	}

	// A user-supplied name already ending in "-1" must be preserved too:
	// collision on "photo-1.txt" must yield "photo-1-1.txt".
	photoName := filepath.Join(tempDir, "photo-1.txt")
	if err := os.WriteFile(photoName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newPhoto, err := FindAvailableNameInc(photoName)
	if err != nil {
		t.Fatalf("FindAvailableNameInc() photo error: %v", err)
	}

	expectedPhoto := filepath.Join(tempDir, "photo-1-1.txt")
	if newPhoto != expectedPhoto {
		t.Errorf("FindAvailableNameInc() = %v; want %v", newPhoto, expectedPhoto)
	}
}

func TestFindAvailableNameIncExhausted(t *testing.T) {
	const maxAttempts = 3
	oldMaxIncrementAttempts := MaxIncrementAttempts
	MaxIncrementAttempts = maxAttempts
	t.Cleanup(func() { MaxIncrementAttempts = oldMaxIncrementAttempts })

	tempDir := t.TempDir()
	baseName := filepath.Join(tempDir, "base.txt")
	if err := os.WriteFile(baseName, []byte("base"), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= maxAttempts; i++ {
		name := filepath.Join(tempDir, fmt.Sprintf("base-%d.txt", i))
		if err := os.WriteFile(name, []byte("conflict"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	got, err := FindAvailableNameInc(baseName)
	if !errors.Is(err, ErrMaxAttemptsReached) {
		t.Errorf("FindAvailableNameInc() error = %v; want %v", err, ErrMaxAttemptsReached)
	}
	if got != "" {
		t.Errorf("FindAvailableNameInc() = %q; want empty string", got)
	}
}

func TestFindAvailableNameTS(t *testing.T) {
	tempDir := t.TempDir()

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create initial file
	if err := os.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test timestamp suffix format
	newName, err := FindAvailableNameTS(baseName)
	if err != nil {
		t.Fatalf("FindAvailableNameTS() error: %v", err)
	}

	// Check that name follows pattern: basename-YYYYMMDD-hhmmss-.ext
	tsPattern := regexp.MustCompile(`-\d{8}-\d{6}\.\d{9}\.txt$`)
	if !tsPattern.MatchString(newName) {
		t.Errorf("FindAvailableNameTS() = %v; want timestamp suffix matching pattern %v",
			newName, tsPattern)
	}

	// Create file with first timestamp and test second timestamp
	if err := os.WriteFile(newName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName2, err := FindAvailableNameTS(baseName)
	if err != nil {
		t.Fatalf("FindAvailableNameTS() second call error: %v, firstname: %v", err, newName)
	}

	// Verify the second name follows the timestamp pattern
	if !tsPattern.MatchString(newName2) {
		t.Errorf("FindAvailableNameTS() = %v; want timestamp suffix matching pattern %v", newName2, tsPattern)
	}

	// Test file without extension
	baseNameNoExt := filepath.Join(tempDir, "testfile")
	if err := os.WriteFile(baseNameNoExt, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName3, err := FindAvailableNameTS(baseNameNoExt)
	if err != nil {
		t.Fatalf("FindAvailableNameTS() no extension error: %v", err)
	}

	// Check timestamp suffix without extension
	tsPatternNoExt := regexp.MustCompile(`-\d{8}-\d{6}\.\d{9}$`)
	if !tsPatternNoExt.MatchString(newName3) {
		t.Errorf("FindAvailableNameTS() = %v; want timestamp suffix without extension", newName3)
	}

	// Preserve the complete original name, including a trailing numeric suffix.
	preservedBaseName := filepath.Join(tempDir, "myfile-2024.txt")
	if err := os.WriteFile(preservedBaseName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	preservedName, err := FindAvailableNameTS(preservedBaseName)
	if err != nil {
		t.Fatalf("FindAvailableNameTS() preserved name error: %v", err)
	}

	preservedPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(filepath.Join(tempDir, "myfile-2024-")) + `\d{8}-\d{6}\.\d{9}\.txt$`)
	if !preservedPattern.MatchString(preservedName) {
		t.Errorf("FindAvailableNameTS() = %v; want full original name preserved with timestamp", preservedName)
	}
}
