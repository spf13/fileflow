## 2024-05-18 - Avoid buffered IO for file-to-file copies
**Learning:** Wrapping `os.File` in `bufio.Writer` before calling `io.Copy` breaks kernel-level zero-copy optimizations (like `copy_file_range` or `sendfile`). Go's standard library `io.Copy` automatically uses these fast syscalls when copying directly between two `os.File` descriptors.
**Action:** Always pass raw `os.File` pointers directly to `io.Copy` when copying files, and avoid intermediate user-space buffers like `bufio`.
