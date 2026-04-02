## 2024-05-24 - Zero-Copy File Copy Optimization
**Learning:** Wrapping `*os.File` in `bufio.Writer` disables Go's internal zero-copy optimizations (like `sendfile` or `copy_file_range`) within `io.Copy`, significantly degrading file copy performance and causing excessive memory allocations.
**Action:** Avoid wrapping standard file descriptors in `bufio.Writer` when performing `io.Copy` operations to leverage the OS-level zero-copy optimizations.
