package fileflow

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIdenticalFiles(t *testing.T) {
	// Set up test files
	file1 := "test_file1.txt"
	file2 := "test_file2.txt"
	file3 := "test_file3.txt"

	content := []byte("Hello World")
	diffContent := []byte("Hello World!")

	// Create test files
	if err := ioutil.WriteFile(file1, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file1)

	if err := ioutil.WriteFile(file2, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file2)

	if err := ioutil.WriteFile(file3, diffContent, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file3)

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

func TestCopy(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "test_copyfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := ioutil.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	if err := Copy(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile() error: %v", err)
	}

	dstContent, err := ioutil.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Reading destination file error: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination file content = %s; want %s", dstContent, content)
	}
}

func TestMoveFile(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_safemove")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := ioutil.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	finalDst, err := Move(srcPath, dstPath)
	if err != nil {
		t.Fatalf("SafeMoveFile() error: %v", err)
	}

	if finalDst != dstPath {
		t.Errorf("SafeMoveFile() = %v; want %v", finalDst, dstPath)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Errorf("Source file still exists after move")
	}

	dstContent, err := ioutil.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Reading destination file error: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination file content = %s; want %s", dstContent, content)
	}
}

func TestExists(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_fileexists")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	existingFile := filepath.Join(tempDir, "exists.txt")
	nonExistingFile := filepath.Join(tempDir, "does_not_exist.txt")

	if err := ioutil.WriteFile(existingFile, []byte("Hello"), 0644); err != nil {
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

func TestMove(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_safemoveefficient")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")
	content := []byte("Hello World")

	if err := ioutil.WriteFile(srcPath, content, 0644); err != nil {
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

	dstContent, err := ioutil.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Reading destination file error: %v", err)
	}

	if !bytes.Equal(content, dstContent) {
		t.Errorf("Destination file content = %s; want %s", dstContent, content)
	}
}
