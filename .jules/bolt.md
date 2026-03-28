## 2024-03-28 - Zero-Copy File Transfers
**Learning:** In Go, wrapping `os.File` in `bufio.Writer` disables zero-copy system calls like `sendfile` or `copy_file_range` during `io.Copy`, which degrades performance for file operations.
**Action:** Always use raw `os.File` objects with `io.Copy` unless there's a specific requirement for buffering.
