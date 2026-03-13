package fileflow

import (
	"bufio"
	"io"
	"os"
	"testing"
)

func BenchmarkCopyOSFile(b *testing.B) {
	srcName := "src_os.txt"
	dstName := "dst_os.txt"
	content := make([]byte, 100*1024*1024) // 100MB
	os.WriteFile(srcName, content, 0644)
	defer os.Remove(srcName)
	defer os.Remove(dstName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src, _ := os.Open(srcName)
		dst, _ := os.Create(dstName)

		io.Copy(dst, src)

		src.Close()
		dst.Close()
	}
}

func BenchmarkCopyBufio(b *testing.B) {
	srcName := "src_buf.txt"
	dstName := "dst_buf.txt"
	content := make([]byte, 100*1024*1024) // 100MB
	os.WriteFile(srcName, content, 0644)
	defer os.Remove(srcName)
	defer os.Remove(dstName)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		src, _ := os.Open(srcName)
		dst, _ := os.Create(dstName)

		writer := bufio.NewWriterSize(dst, 32*1024)
		io.Copy(writer, src)
		writer.Flush()

		src.Close()
		dst.Close()
	}
}
