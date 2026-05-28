## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-28 - sync.Pool and mutable slice capacity
**Learning:** When using `sync.Pool` to reuse `[]byte` buffers whose required size relies on a mutable global variable (`BufferSize`), it's essential to verify the retrieved slice capacity (`cap(*ptr) < size`) before use to prevent bounds errors if the global variable changes between calls.
**Action:** Always verify capacity after retrieving a slice from `sync.Pool` if the needed size is dynamic/mutable, and read the mutable global once into a local variable before the check to avoid TOCTOU race conditions.
