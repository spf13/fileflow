## 2024-03-26 - Enable Zero-Copy in io.Copy
**Learning:** In Go, wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and significantly degrades file copy performance.
**Action:** Do not use `bufio.Writer` around `*os.File` when copying files with `io.Copy`; pass the `*os.File` directly to enable fast path optimizations.
