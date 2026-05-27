## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-27 - Use sync.Pool for file buffer reuse
**Learning:** Using `sync.Pool` to reuse `[]byte` buffers for operations like `Equal` and `Copy` significantly reduces memory allocations and GC pressure, leading to faster execution times. When storing `[]byte` in `sync.Pool`, store pointers `*[]byte` to avoid the heap allocation caused by converting `[]byte` to `interface{}`. Ensure to check slice capacity against mutable `BufferSize` at retrieval time.
**Action:** Always use `sync.Pool` with pointers for frequently allocated large buffers in I/O operations.
