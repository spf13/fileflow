## 2024-03-21 - [Zero-Copy Optimization for io.Copy]
**Learning:** Wrapping standard file objects (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and significantly degrades file copy performance by adding unnecessary memory allocations and overhead.
**Action:** Do not wrap standard files (`*os.File`) in `bufio.Writer` when using `io.Copy`. Pass the `*os.File` directly to `io.Copy` to take advantage of zero-copy optimizations.
