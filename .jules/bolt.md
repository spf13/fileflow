## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-15 - Reduce Buffer Allocations with sync.Pool
**Learning:** Allocating a new 32KB buffer for every file comparison (`Equal`) or copy operation (`Copy`) via `make([]byte, BufferSize)` causes significant memory allocation overhead during frequent file operations.
**Action:** Use a `sync.Pool` to reuse `[]byte` buffers across concurrent operations, drastically reducing memory allocations and improving throughput.
