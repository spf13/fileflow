package fileflow

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkCopyCurrent(b *testing.B) {
	content := make([]byte, 10*1024*1024) // 10MB file
	src := filepath.Join(b.TempDir(), "src.bin")
	ioutil.WriteFile(src, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := filepath.Join(b.TempDir(), "dst.bin")
		sourceFile, _ := os.Open(src)
		destFile, _ := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		writer := bufio.NewWriterSize(destFile, BufferSize)
		io.Copy(writer, sourceFile)
		writer.Flush()
		destFile.Sync()
		destFile.Close()
		sourceFile.Close()
	}
}

func BenchmarkCopyOptimized(b *testing.B) {
	content := make([]byte, 10*1024*1024) // 10MB file
	src := filepath.Join(b.TempDir(), "src.bin")
	ioutil.WriteFile(src, content, 0644)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dst := filepath.Join(b.TempDir(), "dst.bin")
		sourceFile, _ := os.Open(src)
		destFile, _ := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
		io.Copy(destFile, sourceFile) // Direct copy
		destFile.Sync()
		destFile.Close()
		sourceFile.Close()
	}
}
