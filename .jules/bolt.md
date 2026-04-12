## 2024-04-12 - Zero-Copy optimization for file copying
**Learning:** Wrapping `*os.File` in `bufio.Writer` when using `io.Copy` disables Go's ability to use zero-copy system calls like `sendfile` or `copy_file_range`, leading to manual user-space memory copying, more allocations, and slower performance.
**Action:** Always pass raw `*os.File` handles directly to `io.Copy` when copying between files on disk.
