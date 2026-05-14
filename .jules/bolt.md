## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-14 - sync.Pool for byte slices
**Learning:** In Go, when pooling `[]byte` slices to avoid allocations using `sync.Pool`, we must store `*[]byte` instead of `[]byte` to prevent heap allocations during the `interface{}` conversion required by `Put()`. We must also cap check slices fetched from the pool against the dynamic `BufferSize` to resize if needed.
**Action:** Use `*[]byte` in `sync.Pool` implementations for dynamic buffers, and always check capacity upon retrieval.
