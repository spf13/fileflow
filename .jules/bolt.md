## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2026-05-08 - Use sync.Pool for buffer slices in Go
**Learning:** When allocating large `[]byte` slices inside frequently called functions (like file copying or checking equality), it causes unnecessary heap allocations and GC pressure. Storing `*[]byte` instead of `[]byte` in a `sync.Pool` prevents the pool from allocating memory due to interface conversion.
**Action:** Always use `sync.Pool` with pointers to slices (`*[]byte`) to reuse large buffers across file operations, reducing memory usage and garbage collection overhead. Verify slice capacity when fetching from the pool if the required size is mutable.
