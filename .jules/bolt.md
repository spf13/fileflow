## 2024-04-19 - Avoid bufio.Writer when using io.Copy on files
**Learning:** Wrapping a standard `*os.File` in a `bufio.Writer` when using `io.Copy` disables Go's zero-copy system call optimizations (like `sendfile` or `copy_file_range`). The buffer prevents the fast path in `io.Copy` from recognizing both reader and writer as files, leading to userspace memory copying and significantly lower performance.
**Action:** Use `io.Copy(destFile, sourceFile)` directly without `bufio.Writer` for file-to-file copies to leverage kernel-level zero-copy operations.
