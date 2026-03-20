## 2024-05-24 - Zero-Copy File Copying
**Learning:** In Go, wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** When implementing file copying functions, use `io.Copy` directly with the `*os.File` objects rather than using a `bufio.Writer`.
