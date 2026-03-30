## 2024-03-30 - Remove bufio.Writer from file copies
**Learning:** In this codebase, avoid wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy`, as it disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** Always pass `*os.File` directly to `io.Copy` to allow the Go runtime to utilize efficient zero-copy syscalls.
