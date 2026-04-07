## 2024-04-07 - Zero-Copy File Transfers
**Learning:** Wrapping an `*os.File` with `bufio.Writer` before passing it to `io.Copy` disables zero-copy system calls (like `copy_file_range` and `sendfile`). Go's `io.Copy` detects when both source and destination are `*os.File` types and triggers these optimizations, but the `bufio.Writer` wrap prevents this type assertion.
**Action:** Always copy directly between `*os.File` descriptors when possible to leverage system-level zero-copy optimizations.
