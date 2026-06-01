## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-24 - Use sync.Pool for slice buffers efficiently
**Learning:** When using `sync.Pool` to reuse slices whose required size may change at runtime (e.g., dependent on a mutable global variable like `BufferSize`), storing and retrieving pointers to the slice (`*[]byte`) instead of the slice directly avoids interface conversion allocations. Always read the mutable global variable once into a local variable before checking capacity and explicitly reslicing to prevent race conditions.
**Action:** Store `*[]byte` in `sync.Pool`, read dynamic size into a local variable before checking capacity, and explicitly reslice before returning the buffer for use.
