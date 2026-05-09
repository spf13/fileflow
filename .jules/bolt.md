## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-09 - Use sync.Pool for large buffer slices
**Learning:** When using `sync.Pool` to manage `[]byte` slices in Go to avoid allocations, you must store and retrieve pointers to the slice (`*[]byte`) instead of the slice directly. Converting a `[]byte` to an `interface{}` when putting it into the pool natively causes an additional heap allocation.
**Action:** Always pool `*[]byte` instead of `[]byte`, and verify slice length or capacity matches your required size when retrieving it, reallocating if needed.
