## 2024-04-15 - [File Copy Performance: zero-copy vs bufio]
**Learning:** Wrapping an `*os.File` in `bufio.Writer` when using `io.Copy` in Go actually disables zero-copy system calls (like `sendfile` and `copy_file_range`). The overhead of buffering negates the performance benefits of zero-copy syscalls for large file transfers.
**Action:** Avoid wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy`, as it disables zero-copy system calls and degrades file copy performance.
