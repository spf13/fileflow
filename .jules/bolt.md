## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-06-18 - sync.Pool in Go for temporary buffers
**Learning:** In Go, memory allocation and garbage collection for short-lived slices, such as buffers used during `io.CopyBuffer` or file read operations, can be a major performance bottleneck, especially for frequent file operations. `sync.Pool` enables reusing these allocations across operations. When using `sync.Pool` to avoid allocations by pooling `[]byte` slices, storing pointers to the slice (`*[]byte`) instead of the slice directly prevents additional heap allocation when converting `[]byte` to an `interface{}` to put it into the pool. Also, because required slice capacity may change at runtime, checking capacity (e.g. `cap(b) < size`) before reusing and reallocating if needed is necessary.
**Action:** Use `sync.Pool` to pool pointers to `[]byte` slices for temporary buffer allocations in I/O operations to significantly reduce memory allocations and improve throughput.
