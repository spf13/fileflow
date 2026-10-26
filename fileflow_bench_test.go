package fileflow

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCopy(b *testing.B) {
	tmpDir := b.TempDir()
	src := filepath.Join(tmpDir, "src.bin")
	dst := filepath.Join(tmpDir, "dst.bin")

	const size = 10 * 1024 * 1024
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.SetBytes(size)
	b.ResetTimer()
	for b.Loop() {
		final, err := Copy(src, dst)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(final)
	}
}

func BenchmarkEqual(b *testing.B) {
	tmpDir := b.TempDir()
	f1 := filepath.Join(tmpDir, "a.bin")
	f2 := filepath.Join(tmpDir, "b.bin")
	const size = 1 * 1024 * 1024
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(f1, data, 0644); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(f2, data, 0644); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(size)
	b.ResetTimer()
	for b.Loop() {
		if _, err := Equal(f1, f2); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReplace(b *testing.B) {
	tmpDir := b.TempDir()
	dst := filepath.Join(tmpDir, "dst.bin")
	const size = 1 * 1024 * 1024
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(size)
	b.ResetTimer()
	for b.Loop() {
		src := filepath.Join(tmpDir, "src.bin")
		if err := os.WriteFile(src, data, 0644); err != nil {
			b.Fatal(err)
		}
		if _, err := Replace(src, dst); err != nil {
			b.Fatal(err)
		}
	}
}
