## 2024-05-24 - [Avoid wrapping *os.File in bufio.Writer when using io.Copy]
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** When copying between files in Go, use `io.Copy(destFile, sourceFile)` directly.
