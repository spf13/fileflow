## 2024-04-16 - Removed bufio wrapping to enable zero-copy file copy
**Learning:** In this codebase, wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance. `io.Copy` handles `*os.File` efficiently directly.
**Action:** Do not use `bufio.Writer` with `io.Copy` for standard files; pass the files directly.
