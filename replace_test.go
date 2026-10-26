package fileflow

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// A conflict-safe rename must not be substituted here: the canonical path must
// contain the new bytes, without a suffixed file or a remaining source.
func TestReplaceCanonicalPath(t *testing.T) {
	for _, existing := range []bool{false, true} {
		name := "new"
		if existing {
			name = "existing"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "staged.toml")
			dst := filepath.Join(dir, "kennyb.toml")
			writeReplacementFixture(t, src, "new deck")
			if existing {
				writeReplacementFixture(t, dst, "old deck")
			}
			final, err := Replace(src, dst)
			if err != nil || final != dst {
				t.Fatalf("Replace = %q, %v; want %q, nil", final, err, dst)
			}
			assertReplacementContents(t, dst, "new deck")
			if _, err := os.Stat(src); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("source still present: %v", err)
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 || entries[0].Name() != "kennyb.toml" {
				t.Fatalf("unexpected files after replace: %v", entries)
			}
		})
	}
}

func TestReplaceDoesNotUseConflictNaming(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "stage"), filepath.Join(dir, "deck")
	writeReplacementFixture(t, src, "new deck")
	writeReplacementFixture(t, dst, "old deck")
	flow := Flow{FindAvailableName: func(string) (string, error) {
		return "", errors.New("conflict naming must not redirect a replacement")
	}}
	if _, err := flow.Replace(src, dst); err != nil {
		t.Fatal(err)
	}
	assertReplacementContents(t, dst, "new deck")
}

func TestReplaceFailurePreservesDestination(t *testing.T) {
	for _, kind := range []string{"missing source", "directory source", "symlink source"} {
		t.Run(kind, func(t *testing.T) {
			dir := t.TempDir()
			src, dst := filepath.Join(dir, "stage"), filepath.Join(dir, "deck")
			writeReplacementFixture(t, dst, "old deck")
			switch kind {
			case "directory source":
				if err := os.Mkdir(src, 0700); err != nil {
					t.Fatal(err)
				}
			case "symlink source":
				target := filepath.Join(dir, "target")
				writeReplacementFixture(t, target, "source target")
				if err := os.Symlink(target, src); err != nil {
					t.Skipf("symlinks unavailable: %v", err)
				}
			}
			final, err := Replace(src, dst)
			if err == nil || final != "" {
				t.Fatalf("Replace = %q, %v; want empty path and error", final, err)
			}
			if kind == "missing source" && !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("missing source error = %v", err)
			}
			if kind != "missing source" && !errors.Is(err, fs.ErrInvalid) {
				t.Fatalf("non-regular source error = %v", err)
			}
			assertReplacementContents(t, dst, "old deck")
		})
	}
}

func TestReplaceSameFilePreservesSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "deck")
	writeReplacementFixture(t, src, "unchanged")
	for _, alias := range []string{src, dir + string(os.PathSeparator) + "." + string(os.PathSeparator) + "deck"} {
		if _, err := Replace(src, alias); !errors.Is(err, ErrSameFile) {
			t.Fatalf("same file error = %v", err)
		}
		assertReplacementContents(t, src, "unchanged")
	}
	hardlink := filepath.Join(dir, "alias")
	if err := os.Link(src, hardlink); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}
	if _, err := Replace(src, hardlink); !errors.Is(err, ErrSameFile) {
		t.Fatalf("hard link error = %v", err)
	}
	assertReplacementContents(t, src, "unchanged")
	assertReplacementContents(t, hardlink, "unchanged")
}

func TestReplaceRespectsRequiredDestinationDirectory(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "stage")
	dst := filepath.Join(dir, "required-mount", "deck")
	writeReplacementFixture(t, src, "new deck")
	flow := Flow{NoCreateDirs: true}
	if _, err := flow.Replace(src, dst); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing destination directory error = %v", err)
	}
	assertReplacementContents(t, src, "new deck")
	if _, err := os.Stat(filepath.Dir(dst)); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("required directory was created: %v", err)
	}
	if _, err := Replace(src, dst); err != nil {
		t.Fatal(err)
	}
	assertReplacementContents(t, dst, "new deck")
}

func TestReplaceCannotOverwriteDirectory(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "stage"), filepath.Join(dir, "deck")
	writeReplacementFixture(t, src, "new deck")
	if err := os.Mkdir(dst, 0700); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(dst, "keep")
	writeReplacementFixture(t, child, "old data")
	if final, err := Replace(src, dst); err == nil || final != "" {
		t.Fatalf("Replace directory = %q, %v", final, err)
	}
	assertReplacementContents(t, src, "new deck")
	assertReplacementContents(t, child, "old data")
}

func TestReplaceDestinationSymlinkDoesNotOverwriteTarget(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "stage"), filepath.Join(dir, "deck")
	target := filepath.Join(dir, "target")
	writeReplacementFixture(t, src, "new deck")
	writeReplacementFixture(t, target, "outside data")
	if err := os.Symlink(target, dst); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := Replace(src, dst); err != nil {
		t.Fatal(err)
	}
	assertReplacementContents(t, dst, "new deck")
	assertReplacementContents(t, target, "outside data")
	info, err := os.Lstat(dst)
	if err != nil || !info.Mode().IsRegular() {
		t.Fatalf("destination is not the published regular file: %v, %v", info, err)
	}
}

// Truncating or deleting the destination before publishing would expose partial
// bytes or ENOENT to a reader that repeatedly opens the canonical path.
func TestReplaceReadersSeeCompleteFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows open-file sharing can reject replacement; native replacement is covered separately")
	}
	dir := t.TempDir()
	dst := filepath.Join(dir, "deck")
	a, b := bytes.Repeat([]byte("a"), 65536), bytes.Repeat([]byte("b"), 65536)
	if err := os.WriteFile(dst, a, 0600); err != nil {
		t.Fatal(err)
	}
	started, stop, done := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	go func() {
		close(started)
		for {
			data, err := os.ReadFile(dst)
			if err != nil {
				done <- err
				return
			}
			if !bytes.Equal(data, a) && !bytes.Equal(data, b) {
				done <- errors.New("reader observed incomplete destination")
				return
			}
			select {
			case <-stop:
				done <- nil
				return
			default:
			}
		}
	}()
	<-started
	defer func() {
		close(stop)
		if err := <-done; err != nil {
			t.Error(err)
		}
	}()
	for i := 0; i < 32; i++ {
		src := filepath.Join(dir, "stage")
		data := a
		if i%2 == 0 {
			data = b
		}
		if err := os.WriteFile(src, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Replace(src, dst); err != nil {
			t.Fatal(err)
		}
	}
}

func writeReplacementFixture(t *testing.T, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
}

func assertReplacementContents(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s contains %q, want %q", path, got, want)
	}
}
