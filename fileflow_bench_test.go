package fileflow

import (
	"crypto/rand"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCopy(b *testing.B) {
	tempDir, err := ioutil.TempDir("", "bench_copy")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "source.dat")

	// Create a 10MB file
	content := make([]byte, 10*1024*1024)
	rand.Read(content)
	if err := ioutil.WriteFile(srcPath, content, 0644); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstPath := filepath.Join(tempDir, "dest.dat")
		if err := Copy(srcPath, dstPath); err != nil {
			b.Fatal(err)
		}
		os.Remove(dstPath)
	}
}
