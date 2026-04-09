## 2024-04-09 - Zero-Copy File Transfers
**Learning:** Wrapping `*os.File` in `bufio.Writer` before calling `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) because it hides the underlying file interface, forcing user-space copying which significantly degrades performance and increases memory allocations.
**Action:** Always pass `*os.File` directly to `io.Copy` when copying between standard files to fully leverage OS-level optimizations.
