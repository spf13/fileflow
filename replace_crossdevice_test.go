//go:build linux

package fileflow

import (
	"context"
	"crypto/sha256"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func replacementCrossDeviceSource(t *testing.T, destinationDir string) string {
	t.Helper()
	dir, err := os.MkdirTemp("/dev/shm", "fileflow-replace-test-")
	if err != nil {
		t.Skipf("second filesystem unavailable: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0700) // undo the read-only test directory
		_ = os.RemoveAll(dir)
	})
	a, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.Stat(destinationDir)
	if err != nil {
		t.Fatal(err)
	}
	if a.Sys().(*syscall.Stat_t).Dev == b.Sys().(*syscall.Stat_t).Dev {
		t.Skip("test requires distinct filesystems")
	}
	return dir
}

func TestReplaceCrossFilesystem(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "stage"), filepath.Join(dstDir, "deck")
	writeReplacementFixture(t, src, "new deck")
	writeReplacementFixture(t, dst, "old deck")
	if err := os.Chmod(src, 0640); err != nil {
		t.Fatal(err)
	}
	final, err := Replace(src, dst)
	if err != nil || final != dst {
		t.Fatalf("Replace = %q, %v; want %q, nil", final, err, dst)
	}
	assertReplacementContents(t, dst, "new deck")
	if _, err := os.Stat(src); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("source was not consumed: %v", err)
	}
	info, err := os.Stat(dst)
	if err != nil || info.Mode().Perm() != 0640 {
		t.Fatalf("source permissions not preserved: %v, %v", info, err)
	}
	assertReplacementDirectory(t, dstDir, "deck")
}

func TestMoveCrossFilesystemPreservesTimes(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "stage"), filepath.Join(dstDir, "deck")
	writeTimedFixture(t, src)
	if _, err := Move(src, dst); err != nil {
		t.Fatalf("Move() error: %v", err)
	}
	assertFixtureTimes(t, dst)
}

func TestReplaceCrossFilesystemPreservesTimes(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "stage"), filepath.Join(dstDir, "deck")
	writeTimedFixture(t, src)
	writeReplacementFixture(t, dst, "old deck")
	if _, err := Replace(src, dst); err != nil {
		t.Fatalf("Replace() error: %v", err)
	}
	assertFixtureTimes(t, dst)
}

func TestReplaceCrossFilesystemCopyReadFailure(t *testing.T) {
	// This regular procfs file opens but reading its unmapped zero address fails
	// with EIO, exercising a real io.Copy error without filesystem mocks/hooks.
	const src = "/proc/self/mem"
	info, err := os.Lstat(src)
	if err != nil || !info.Mode().IsRegular() {
		t.Skip("read-error fixture unavailable")
	}
	file, err := os.Open(src)
	if err != nil {
		t.Skipf("cannot open read-error fixture: %v", err)
	}
	var probe [1]byte
	_, readErr := file.Read(probe[:])
	_ = file.Close()
	if !errors.Is(readErr, syscall.EIO) {
		t.Skipf("fixture does not produce EIO: %v", readErr)
	}
	dir := t.TempDir()
	dst := filepath.Join(dir, "deck")
	writeReplacementFixture(t, dst, "old deck")
	final, err := Replace(src, dst)
	if final != "" || !errors.Is(err, syscall.EIO) {
		t.Fatalf("Replace = %q, %v; want empty path and EIO", final, err)
	}
	assertReplacementContents(t, dst, "old deck")
	assertReplacementDirectory(t, dir, "deck")
}

func TestReplaceCrossFilesystemSourceRemovalFailure(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory removal permissions; run as an unprivileged user")
	}
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "stage"), filepath.Join(dstDir, "deck")
	writeReplacementFixture(t, src, "committed deck")
	writeReplacementFixture(t, dst, "old deck")
	if err := os.Chmod(srcDir, 0500); err != nil {
		t.Fatal(err)
	}
	final, err := Replace(src, dst)
	var removalErr *ErrFailedRemovingOriginal
	if final != dst || !errors.As(err, &removalErr) || !errors.Is(err, fs.ErrPermission) {
		t.Fatalf("Replace = %q, %v; want committed path and removal permission error", final, err)
	}
	if removalErr.File != src {
		t.Fatalf("cleanup error identifies %q, want %q", removalErr.File, src)
	}
	assertReplacementContents(t, dst, "committed deck")
	assertReplacementContents(t, src, "committed deck")
	assertReplacementDirectory(t, dstDir, "deck")
}

func assertReplacementDirectory(t *testing.T, dir, want string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != want {
		t.Fatalf("unexpected destination files after replacement: %v", entries)
	}
}

func TestMoveContextCrossFilesystemCancel(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	data := writeSizedFixture(t, src, 3*copyChunkSize+copyChunkSize/2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f := Flow{Progress: func(done, total int64) { cancel() }}
	final, err := f.MoveContext(ctx, src, dst)
	if final != "" || !errors.Is(err, context.Canceled) {
		t.Fatalf("MoveContext = %q, %v; want \"\", context.Canceled", final, err)
	}
	assertNoCopyOutput(t, dst)
	assertContents(t, src, data)
}

func TestMoveContextCrossFilesystemProgress(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	size := int64(2*copyChunkSize + 17)
	data := writeSizedFixture(t, src, int(size))
	var calls int
	var last int64
	f := Flow{Progress: func(done, total int64) {
		calls++
		last = done
		if total != size {
			t.Errorf("Progress total = %d; want %d", total, size)
		}
	}}
	final, err := f.MoveContext(context.Background(), src, dst)
	if err != nil || final != dst {
		t.Fatalf("MoveContext = %q, %v; want %q, nil", final, err, dst)
	}
	if calls < 2 || last != size {
		t.Errorf("Progress calls = %d, last done = %d; want >= 2 calls ending at %d", calls, last, size)
	}
	assertContents(t, dst, data)
	if _, err := os.Stat(src); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("source not removed after move: %v", err)
	}
}

func TestMoveResultCrossFilesystemVerify(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	data := writeSizedFixture(t, src, copyChunkSize+123)
	f := Flow{Verify: true}
	r, err := f.MoveResult(context.Background(), src, dst)
	if err != nil || r.Path != dst {
		t.Fatalf("MoveResult = %+v, %v; want path %q", r, err, dst)
	}
	if r.Digest != sha256.Sum256(data) {
		t.Errorf("Digest = %v; want %x", r.Digest, sha256.Sum256(data))
	}
	assertContents(t, dst, data)
	if _, err := os.Stat(src); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("source not removed after verified move: %v", err)
	}
}

func TestMoveResultCrossFilesystemVerifyMismatch(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	data := writeSizedFixture(t, src, 4096)
	corruptBeforeVerify(t)
	f := Flow{Verify: true}
	r, err := f.MoveResult(context.Background(), src, dst)
	if !errors.Is(err, ErrDigestMismatch) || r.Path != "" {
		t.Fatalf("MoveResult = %+v, %v; want empty path, ErrDigestMismatch", r, err)
	}
	assertNoCopyOutput(t, dst)
	assertContents(t, src, data)
}

func TestReplaceCrossFilesystemVerifyMismatch(t *testing.T) {
	dstDir := t.TempDir()
	srcDir := replacementCrossDeviceSource(t, dstDir)
	src, dst := filepath.Join(srcDir, "stage"), filepath.Join(dstDir, "deck")
	writeReplacementFixture(t, src, "new deck")
	writeReplacementFixture(t, dst, "old deck")
	corruptBeforeVerify(t)
	f := Flow{Verify: true}
	final, err := f.Replace(src, dst)
	if final != "" || !errors.Is(err, ErrDigestMismatch) {
		t.Fatalf("Replace = %q, %v; want empty path, ErrDigestMismatch", final, err)
	}
	assertReplacementContents(t, dst, "old deck")
	assertReplacementContents(t, src, "new deck")
	assertReplacementDirectory(t, dstDir, "deck")
}
