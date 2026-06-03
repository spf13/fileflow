## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2026-06-03 - Use sync.Pool for buffer reuse
**Learning:** In heavily concurrent or repeating file operations (like file copy/equal), repeatedly allocating 32KB byte slices creates unnecessary GC pressure. Reusing these slices via a package-level sync.Pool significantly drops memory allocations per operation. Note: Store pointers to slices (`*[]byte`) in the pool instead of the slice directly to avoid an extra interface conversion allocation on `.Put`, and read the mutable global `BufferSize` to correctly size/reslice before returning from the pool.
**Action:** Use `sync.Pool` to manage large temporary buffers instead of making fresh slices with `make([]byte, ...)` during repeated I/O.
