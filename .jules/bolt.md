## 2024-04-22 - Avoid bufio with io.Copy for Files
**Learning:** Wrapping `os.File` in a `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`), degrading performance and increasing memory allocation.
**Action:** Pass `*os.File` directly to `io.Copy` to allow the Go runtime to utilize platform-specific zero-copy optimizations.
