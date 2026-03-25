## 2024-05-24 - Zero-copy `io.Copy` in Go
**Learning:** In this codebase, avoiding wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` is crucial because it disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** When copying data directly between `*os.File` instances, pass the raw file descriptors to `io.Copy` to allow the Go runtime to leverage OS-level zero-copy optimizations.
