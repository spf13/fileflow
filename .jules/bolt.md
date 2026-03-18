## 2026-03-18 - Unbuffered file copies enable zero-copy syscalls
**Learning:** In Go, wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance. `io.Copy` is already optimized for standard files and uses these zero-copy features implicitly when copying from and to `*os.File`.
**Action:** Always use `io.Copy(dst, src)` directly without wrapping the destination `*os.File` in `bufio.Writer` or `bufio.NewWriterSize` to maximize file copy performance.
