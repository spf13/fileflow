package fileflow

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestIdenticalFiles(t *testing.T) {
	testCases := []struct {
		file1, file2 string
		expected     bool
	}{
		{"file1.txt", "file2.txt", true},
		{"file1.txt", "file1.txt", true},
		{"file1.txt", "file3.txt", false},
	}

	// Create test files
	if err := ioutil.WriteFile("file1.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("file1.txt")

	if err := ioutil.WriteFile("file2.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("file2.txt")

	if err := ioutil.WriteFile("file3.txt", []byte("Hellod World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("file3.txt")

	for _, tc := range testCases {
		result, err := IdenticalFiles(tc.file1, tc.file2)
		if err != nil {
			t.Errorf("IdenticalFiles(%q, %q) unexpected error: %v", tc.file1, tc.file2, err)
			continue
		}
		if result != tc.expected {
			t.Errorf("IdenticalFiles(%q, %q) = %v; want %v", tc.file1, tc.file2, result, tc.expected)
		}
	}
}

func TestCopyFile(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "fileflow_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		src, dst string
		expected bool
	}{
		{"filea.txt", "file2.txt", true},
		{"fileb.txt", "folder/file2.txt", true},
	}

	// Create test files in temp directory
	if err := ioutil.WriteFile(filepath.Join(tempDir, "filea.txt"), []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tempDir, "fileb.txt"), []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		srcPath := filepath.Join(tempDir, tc.src)
		dstPath := filepath.Join(tempDir, tc.dst)

		// Create folder if needed
		if filepath.Dir(dstPath) != tempDir {
			if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
				t.Fatal(err)
			}
		}

		err := CopyFile(ctx, srcPath, dstPath)
		if (err == nil) != tc.expected {
			t.Errorf("CopyFile(%q, %q) = %v; want %v", srcPath, dstPath, (err == nil), tc.expected)
		}
	}
}

func TestSafeMoveFile(t *testing.T) {
	ctx := context.Background()

	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "fileflow_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		src, dst, actualdst string
		expected            bool
		dstExists           bool
	}{
		{"file0.txt", "file0.txt", "", false, false},                      // File same
		{"file1.txt", "file2.txt", "file2.txt", true, true},               // Basic move
		{"fileA.txt", "folder/fileA.txt", "folder/fileA.txt", true, true}, // Basic move with folder
		{"fileB.txt", "fileB-diff.txt", "fileB-diff-1.txt", true, true},   // dst file exists and is different
		{"fileC.txt", "fileC-1.txt", "fileC-1.txt", true, true},           // dst file exists and is the same
		{"fileD.txt", "fileD-1.txt", "fileD-2.txt", true, true},           // dst file exists and is different, and incremented
	}

	// Create test files in temp directory
	for _, tc := range testCases {
		srcPath := filepath.Join(tempDir, tc.src)
		if err := ioutil.WriteFile(srcPath, []byte("Hello World"), 0666); err != nil {
			t.Fatal(err)
		}
	}

	// Create additional test files
	if err := ioutil.WriteFile(filepath.Join(tempDir, "fileB-diff.txt"), []byte("Hello World FOO"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tempDir, "fileC-1.txt"), []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filepath.Join(tempDir, "fileD-1.txt"), []byte("Hello World FOO"), 0666); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		srcPath := filepath.Join(tempDir, tc.src)
		dstPath := filepath.Join(tempDir, tc.dst)
		expectedDstPath := filepath.Join(tempDir, tc.actualdst)

		// Create folder if needed
		if filepath.Dir(dstPath) != tempDir {
			if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
				t.Fatal(err)
			}
		}

		finaldst, err := SafeMoveFile(ctx, srcPath, dstPath)
		if tc.expected && err != nil {
			t.Errorf("SafeMoveFile(%q, %q) unexpected error: %v", srcPath, dstPath, err)
			continue
		}
		if !tc.expected && err == nil {
			t.Errorf("SafeMoveFile(%q, %q) expected error but got none", srcPath, dstPath)
			continue
		}
		if tc.expected && tc.actualdst != "" && finaldst != expectedDstPath {
			t.Errorf("SafeMoveFile(%q, %q) = %v; want %v", srcPath, dstPath, finaldst, expectedDstPath)
		}
		if tc.dstExists && !FileExists(dstPath) {
			t.Errorf("SafeMoveFile(%q, %q) destination file doesn't exist", srcPath, dstPath)
		}
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "fileflow_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		file     string
		expected bool
	}{
		{"file1.txt", true},
		{"fileFoo.txt", false},
		{"folder/file1.txt", false},
	}

	// Create test file
	if err := ioutil.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}

	for _, tc := range testCases {
		path := filepath.Join(tempDir, tc.file)
		result := FileExists(path)
		if result != tc.expected {
			t.Errorf("FileExists(%q) = %v; want %v", path, result, tc.expected)
		}
	}
}

func TestSafeMoveFileEfficient(t *testing.T) {
	ctx := context.Background()
	// Create test directory and files
	testDir, err := ioutil.TempDir("", "fileflow_test_efficient")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	srcContent := []byte("test content")
	srcPath := filepath.Join(testDir, "source.txt")
	dstPath := filepath.Join(testDir, "dest.txt")

	if err := ioutil.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		src     string
		dst     string
		want    string
		wantErr bool
	}{
		{
			name:    "Basic move",
			src:     srcPath,
			dst:     dstPath,
			want:    dstPath,
			wantErr: false,
		},
		{
			name:    "Same file",
			src:     srcPath,
			dst:     srcPath,
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset source file for each test
			if err := ioutil.WriteFile(srcPath, srcContent, 0644); err != nil {
				t.Fatal(err)
			}

			got, err := SafeMoveFileEfficient(ctx, tt.src, tt.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeMoveFileEfficient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeMoveFileEfficient() = %v, want %v", got, tt.want)
			}

			if !tt.wantErr {
				// Verify destination file exists and has correct content
				content, err := ioutil.ReadFile(tt.dst)
				if err != nil {
					t.Errorf("Failed to read destination file: %v", err)
				} else if !bytes.Equal(content, srcContent) {
					t.Errorf("Destination file content = %v, want %v", content, srcContent)
				}
			}
		})
	}
}
