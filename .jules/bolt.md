## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-17 - Avoid interface{} allocations with sync.Pool and handle dynamic sizes
**Learning:** When using `sync.Pool` to avoid allocations for `[]byte` slices, storing them directly causes an `interface{}` allocation. Furthermore, if the required slice size depends on a mutable global variable (like `BufferSize`), stale smaller buffers might be returned from the pool.
**Action:** Store and retrieve pointers (`*[]byte`) in `sync.Pool` to prevent the `interface{}` conversion heap allocation, and always verify slice capacity upon retrieval, reallocating if the required dynamic size has increased.
