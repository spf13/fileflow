## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2026-06-02 - sync.Pool allocation overhead
**Learning:** When using `sync.Pool` to manage `[]byte` slices to avoid allocations, converting a `[]byte` to an `interface{}` when putting it into the pool natively causes an additional heap allocation.
**Action:** Store and retrieve pointers to the slice (`*[]byte`) instead of the slice directly to avoid the allocation.
