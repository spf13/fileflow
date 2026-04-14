## 2024-04-14 - Zero-copy disabled by bufio.Writer
**Learning:** Wrapping an `os.File` in a `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` and `copy_file_range`, unexpectedly degrading performance and increasing memory allocations.
**Action:** Pass `*os.File` directly to `io.Copy` to leverage zero-copy optimizations.
