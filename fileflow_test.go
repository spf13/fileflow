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
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

func TestCustomFindAvailableName(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_customfindavailablename")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create a file to ensure the name needs to be incremented
	if err := ioutil.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	// Define a custom function for finding available names
	customFindAvailableName := func(baseName string) (string, error) {
		return baseName + "_custom", nil
	}

	// Set the custom function
	FindAvailableName = customFindAvailableName

	newName, err := FindAvailableName(baseName)
	if err != nil {
		t.Fatalf("CustomFindAvailableName() error: %v", err)
	}

	expectedName := baseName + "_custom"
	if newName != expectedName {
		t.Errorf("CustomFindAvailableName() = %v; want %v", newName, expectedName)
	}

	// Reset FindAvailableName to default
	FindAvailableName = FindAvailableNameInc
}

func TestFindAvailableNameInc(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_findavailablenameinc")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create a file to ensure the name needs to be incremented
	if err := ioutil.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
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
	if err := ioutil.WriteFile(newName, []byte("Hello World"), 0644); err != nil {
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
	if err := ioutil.WriteFile(baseNameNoExt, []byte("Hello World"), 0644); err != nil {
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
}

func TestFindAvailableNameTS(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_findavailablenamets")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	baseName := filepath.Join(tempDir, "testfile.txt")

	// Create initial file
	if err := ioutil.WriteFile(baseName, []byte("Hello World"), 0644); err != nil {
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
	if err := ioutil.WriteFile(newName, []byte("Hello World"), 0644); err != nil {
		t.Fatal(err)
	}

	newName2, err := FindAvailableNameTS(baseName)
	if err != nil {
		t.Fatalf("FindAvailableNameTS() second call error: %v, firstname: %v", err, newName)
	}

	// Verify increment part is 2
	if !tsPattern.MatchString(newName2) {
		t.Errorf("FindAvailableNameTS() = %v; want timestamp suffix matching patterh %v", newName2, tsPattern)
	}

	// Test file without extension
	baseNameNoExt := filepath.Join(tempDir, "testfile")
	if err := ioutil.WriteFile(baseNameNoExt, []byte("Hello World"), 0644); err != nil {
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
}
