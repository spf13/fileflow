## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-22 - sync.Pool Buffer Allocation Reduction
**Learning:** The fileflow package allocates new 32KB byte buffers repeatedly during operations like Equal and Copy, leading to high garbage collection pressure.
**Action:** Used `sync.Pool` to reuse existing slice allocations. Ensured we store pointers to slices (`*[]byte`) in the pool instead of the slice directly to avoid the implicit heap allocation when converting `[]byte` to `interface{}`.
