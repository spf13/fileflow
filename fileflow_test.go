package fileflow

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewIncrementedName(t *testing.T) {
	testCases := []struct {
		name     string
		i        int
		expected string
	}{
		{"file.txt", 1, "file-1.txt"},
		{"file.txt", 10, "file-10.txt"},
		{"folder/file.txt", 2, "folder/file-2.txt"},
		{"file", 0, "file-0"},
	}

	for _, tc := range testCases {
		result := newIncrementedName(tc.name, tc.i)
		if result != tc.expected {
			t.Errorf("newIncrementedName(%q, %d) = %q; want %q", tc.name, tc.i, result, tc.expected)
		}
	}
}

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
		result := IdenticalFiles(tc.file1, tc.file2)
		if result != tc.expected {
			t.Errorf("IdenticalFiles(%q, %q) = %v; want %v", tc.file1, tc.file2, result, tc.expected)
		}
	}

	os.Remove("file1.txt")
	os.Remove("file2.txt")
	os.Remove("file3.txt")
}

func TestMoveFile(t *testing.T) {
	testCases := []struct {
		src, dst string
		expected bool
	}{
		{"filea.txt", "file2.txt", true},
		{"fileb.txt", "folder/file2.txt", true},
	}

	// Create test files
	if err := ioutil.WriteFile("filea.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("filea.txt")
	if err := ioutil.WriteFile("fileb.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fileb.txt")

	for _, tc := range testCases {
		os.Mkdir("folder", os.ModePerm)
		err := MoveFile(tc.src, tc.dst)
		if (err == nil) != tc.expected {
			t.Errorf("MoveFile(%q, %q) = %v; want %v", tc.src, tc.dst, (err == nil), tc.expected)
		}
		os.Remove(tc.dst)
	}
	os.RemoveAll("folder")
}

func TestSafeMoveFile(t *testing.T) {
	testCases := []struct {
		src, dst, actualdst string
		expected            bool
		dstExists           bool
	}{
		{"file0.txt", "file0.txt", "", false, false},                      // File same
		{"file1.txt", "file2.txt", "file2.txt", true, true},               // Basic move
		{"fileA.txt", "folder/fileA.txt", "folder/fileA.txt", true, true}, // Basic move with folder
		{"fileB.txt", "fileB-diff.txt", "fileB-diff-1.txt", false, true},  // dst file exists and is different
		{"fileC.txt", "fileC-1.txt", "fileC-1.txt", false, true},          // dst file exists and is the same
		{"fileD.txt", "fileD-1.txt", "fileD-2.txt", false, true},          // dst file exists and is different, and incremented
	}

	// Create test files
	for _, tc := range testCases {
		if err := ioutil.WriteFile(tc.src, []byte("Hello World"), 0666); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tc.src)
	}

	if err := ioutil.WriteFile("fileB-diff.txt", []byte("Hello World FOO"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fileB-diff.txt")
	defer os.Remove("fileB-diff-1.txt")

	if err := ioutil.WriteFile("fileC-1.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fileC-1.txt")
	if err := ioutil.WriteFile("fileD.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fileD.txt")
	if err := ioutil.WriteFile("fileD-1.txt", []byte("Hello World FOO"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("fileD-1.txt")
	defer os.Remove("fileD-2.txt")

	for _, tc := range testCases {
		os.Mkdir("folder", os.ModePerm)
		finaldst, err := SafeMoveFile(tc.src, tc.dst)

		if (err == nil) != tc.expected && finaldst != tc.actualdst {
			t.Errorf("SafeMoveFile(%q, %q) = %v, %v; want %v, %v", tc.src, tc.dst, finaldst, (err == nil), tc.actualdst, tc.expected)
		}

		if tc.dstExists {
			if !FileExists(tc.dst) {
				t.Errorf("SafeMoveFile(%q, %q) = %v, %v; want %v, %v", tc.src, tc.dst, finaldst, (err == nil), tc.actualdst, tc.expected)
			}
		}

		defer os.Remove(tc.dst)
	}
	os.RemoveAll("folder")
}

func TestFileExists(t *testing.T) {
	testCases := []struct {
		file     string
		expected bool
	}{
		{"file1.txt", true},
		{"fileFoo.txt", false},
		{"folder/file1.txt", false},
	}

	// Create test files
	if err := ioutil.WriteFile("file1.txt", []byte("Hello World"), 0666); err != nil {
		t.Fatal(err)
	}
	defer os.Remove("file1.txt")

	for _, tc := range testCases {
		result := FileExists(tc.file)
		if result != tc.expected {
			t.Errorf("FileExists(%q) = %v; want %v", tc.file, result, tc.expected)
		}
	}
}

func Test_newIncrementedName(t *testing.T) {
	type args struct {
		name string
		i    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"file.txt", args{"file.txt", 1}, "file-1.txt"},
		{"file.txt", args{"file.txt", 2}, "file-2.txt"},
		{"file.txt", args{"file-1.txt", 3}, "file-3.txt"},
		{"file.txt", args{"file-1-2.txt", 3}, "file-1-3.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newIncrementedName(tt.args.name, tt.args.i); got != tt.want {
				t.Errorf("newIncrementedName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeMoveFileEfficient(t *testing.T) {
	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SafeMoveFileEfficient(tt.args.src, tt.args.dst)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeMoveFileEfficient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SafeMoveFileEfficient() = %v, want %v", got, tt.want)
			}
		})
	}
}
