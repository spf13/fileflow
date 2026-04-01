## 2024-05-24 - Zero-copy Optimization Anti-Pattern
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** Always pass `*os.File` directly to `io.Copy` instead of wrapping it in `bufio` to leverage zero-copy optimizations.
