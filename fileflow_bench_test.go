package fileflow

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCopy(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "fileflow-bench-*")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

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
		err := Copy(src, dst)
		if err != nil {
			b.Fatal(err)
		}
		os.Remove(dst)
	}
}
