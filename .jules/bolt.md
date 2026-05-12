## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-24 - Sync Pool for Buffer Reuse
**Learning:** Using `sync.Pool` to reuse `[]byte` buffers for large buffer sizes (e.g., 32KB) significantly reduces memory allocations and GC pressure in file copying operations.
**Action:** Implement a package-level `sync.Pool` to reuse large `[]byte` buffers for `Equal` and `Copy` functions.
