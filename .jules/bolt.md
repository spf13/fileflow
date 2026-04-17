## 2024-04-17 - Optimize io.Copy for OS Files
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables Go's internal zero-copy optimizations (like `sendfile` or `copy_file_range`) and severely degrades file copy performance and increases memory allocations.
**Action:** Avoid wrapping `*os.File` arguments in `bufio` readers/writers when passing them directly to `io.Copy` to allow the standard library to utilize efficient OS-level syscalls.
