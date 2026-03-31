## 2024-05-24 - Zero-copy File Transfers in Go
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`), degrading file copy performance and causing unnecessary memory allocations.
**Action:** Avoid wrapping `*os.File` in `bufio.Writer` or `bufio.Reader` when performing file-to-file copies using `io.Copy`.
