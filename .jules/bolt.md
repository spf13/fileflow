## 2024-04-13 - Direct io.Copy for Zero-Copy Optimization
**Learning:** Wrapping standard `*os.File` objects with `bufio.Writer` before using `io.Copy` disables Go's ability to use zero-copy system calls (like `sendfile` or `copy_file_range`).
**Action:** Avoid `bufio` when copying directly between files to maximize performance.
