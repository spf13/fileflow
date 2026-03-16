## 2024-05-20 - File Copy Zero-Copy Fast Path
**Learning:** In Go, wrapping an `*os.File` in a `bufio.Writer` before calling `io.Copy` actually degrades performance for file-to-file copying. `io.Copy` has a fast path using `io.ReaderFrom` that delegates to efficient system calls like `sendfile` or `copy_file_range` when both the reader and writer are standard files, entirely bypassing user space memory copying.
**Action:** When copying data between two standard files (`*os.File`), pass the file descriptors directly to `io.Copy` instead of buffering them.
