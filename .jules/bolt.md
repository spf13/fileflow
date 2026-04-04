## 2024-04-04 - Zero-copy optimization in file copying
**Learning:** In Go, wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance, because Go can no longer detect that the destination is a raw file descriptor.
**Action:** Avoid wrapping `*os.File` in `bufio` during `io.Copy` operations to preserve zero-copy optimizations.
