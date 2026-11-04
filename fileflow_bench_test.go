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

	// Create a 10MB test file
	data := make([]byte, 10*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile(src, data, 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Copy(src, dst)
		if err != nil {
			b.Fatal(err)
		}
		if err := os.Remove(dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEqual(b *testing.B) {
	dir := b.TempDir()
	data := make([]byte, 10*1024*1024)
	if _, err := rand.Read(data); err != nil {
		b.Fatal(err)
	}
	f1, f2 := filepath.Join(dir, "a"), filepath.Join(dir, "b")
	for _, f := range []string{f1, f2} {
		if err := os.WriteFile(f, data, 0644); err != nil {
			b.Fatal(err)
		}
	}
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if same, err := Equal(f1, f2); err != nil || !same {
			b.Fatalf("Equal = %v, %v", same, err)
		}
	}
}
