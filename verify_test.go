package fileflow

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// corruptBeforeVerify makes the verification re-read see different bytes.
func corruptBeforeVerify(t *testing.T) {
	t.Helper()
	testHookBeforeVerify = func(tmpName string) {
		if err := os.WriteFile(tmpName, []byte("corrupted"), 0600); err != nil {
			t.Errorf("corrupting temp file: %v", err)
		}
	}
	t.Cleanup(func() { testHookBeforeVerify = nil })
}

func assertNoTemps(t *testing.T, dir string) {
	t.Helper()
	temps, err := filepath.Glob(filepath.Join(dir, ".fileflow-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temps) != 0 {
		t.Errorf("temporary files left: %v", temps)
	}
}

func TestDigestString(t *testing.T) {
	data := []byte("hello, fileflow")
	sum := sha256.Sum256(data)
	d := Digest(sum)
	if got, want := d.String(), hex.EncodeToString(sum[:]); got != want {
		t.Errorf("Digest.String() = %q; want %q", got, want)
	}
	var _ [32]byte = [32]byte(d)
}

func TestCopyResultVerify(t *testing.T) {
	srcDir, dstDir := t.TempDir(), t.TempDir()
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	data := writeSizedFixture(t, src, 2*copyChunkSize+5)

	f := Flow{Verify: true}
	r, err := f.CopyResult(context.Background(), src, dst)
	if err != nil || r.Path != dst {
		t.Fatalf("CopyResult = %+v, %v; want path %q", r, err, dst)
	}
	if r.Digest != sha256.Sum256(data) {
		t.Errorf("Digest = %v; want %x", r.Digest, sha256.Sum256(data))
	}
	assertContents(t, dst, data)
	assertNoTemps(t, dstDir)

	var plain Flow
	r, err = plain.CopyResult(context.Background(), src, filepath.Join(dstDir, "plain"))
	if err != nil {
		t.Fatal(err)
	}
	if r.Digest != (Digest{}) {
		t.Errorf("Digest without Verify = %v; want zero", r.Digest)
	}
}

func TestCopyResultVerifyMismatch(t *testing.T) {
	srcDir, dstDir := t.TempDir(), t.TempDir()
	src, dst := filepath.Join(srcDir, "src"), filepath.Join(dstDir, "dst")
	data := writeSizedFixture(t, src, 4096)
	corruptBeforeVerify(t)

	f := Flow{Verify: true}
	r, err := f.CopyResult(context.Background(), src, dst)
	if !errors.Is(err, ErrDigestMismatch) || r.Path != "" {
		t.Fatalf("CopyResult = %+v, %v; want empty path, ErrDigestMismatch", r, err)
	}
	var moveErr *ErrFailedMovingFile
	if !errors.As(err, &moveErr) {
		t.Errorf("error %v is not *ErrFailedMovingFile", err)
	}
	assertNoCopyOutput(t, dst)
	assertContents(t, src, data)
}

func TestMoveResultSameFSVerify(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "src"), filepath.Join(dir, "sub", "dst")
	data := writeSizedFixture(t, src, 70000)
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
		t.Errorf("source not removed: %v", err)
	}
}

func TestVerifyIdenticalDestination(t *testing.T) {
	ops := map[string]func(*Flow, context.Context, string, string) (Result, error){
		"CopyResult": (*Flow).CopyResult,
		"MoveResult": (*Flow).MoveResult,
	}
	for name, op := range ops {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			src, dst := filepath.Join(dir, "src"), filepath.Join(dir, "dst")
			data := writeSizedFixture(t, src, 9000)
			if err := os.WriteFile(dst, data, 0644); err != nil {
				t.Fatal(err)
			}
			f := Flow{Verify: true}
			r, err := op(&f, context.Background(), src, dst)
			if err != nil || r.Path != dst {
				t.Fatalf("%s = %+v, %v; want path %q", name, r, err, dst)
			}
			if r.Digest != sha256.Sum256(data) {
				t.Errorf("Digest = %v; want %x", r.Digest, sha256.Sum256(data))
			}
			assertNoTemps(t, dir)
			_, statErr := os.Stat(src)
			if name == "MoveResult" && !errors.Is(statErr, fs.ErrNotExist) {
				t.Errorf("source not removed after dedup: %v", statErr)
			}
			if name == "CopyResult" && statErr != nil {
				t.Errorf("source missing after copy: %v", statErr)
			}
		})
	}
}

func TestVerifyConflictPath(t *testing.T) {
	dir := t.TempDir()
	src, dst := filepath.Join(dir, "src.txt"), filepath.Join(dir, "dst.txt")
	data := writeSizedFixture(t, src, 5000)
	if err := os.WriteFile(dst, []byte("other"), 0644); err != nil {
		t.Fatal(err)
	}
	f := Flow{Verify: true}
	r, err := f.CopyResult(context.Background(), src, dst)
	want := filepath.Join(dir, "dst-1.txt")
	if err != nil || r.Path != want {
		t.Fatalf("CopyResult = %+v, %v; want path %q", r, err, want)
	}
	if r.Digest != sha256.Sum256(data) {
		t.Errorf("Digest = %v; want %x", r.Digest, sha256.Sum256(data))
	}
	assertContents(t, want, data)
}
