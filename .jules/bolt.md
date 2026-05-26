## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-26 - Buffer Pool for File Copy
**Learning:** In Go, converting a `[]byte` to an `interface{}` when putting it into a `sync.Pool` natively causes an additional heap allocation. Furthermore, `sync.Pool` needs to be size-aware when dealing with mutable global variables like `BufferSize`.
**Action:** Store and retrieve pointers to the slice (`*[]byte`) instead of the slice directly in the `sync.Pool`. Also verify slice capacity against `BufferSize` when pulling from the pool to avoid buffer overflows if `BufferSize` was changed globally by the user.
