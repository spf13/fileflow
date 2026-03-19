## 2024-05-24 - [Avoid wrapping *os.File in bufio.Writer when using io.Copy]
**Learning:** Wrapping an `*os.File` in `bufio.Writer` and passing it to `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and downgrades it to a standard user-space copy (reading chunks into a buffer and writing them out). Go's `io.Copy` is already highly optimized for `*os.File` to `*os.File` copies.
**Action:** Do not wrap standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` in Go, as it degrades file copy performance.
