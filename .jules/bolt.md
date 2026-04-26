## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-04-26 - Use sync.Pool for large file buffers
**Learning:** In the fileflow package, allocating a new large buffer (e.g. 32KB) for every file comparison (`Equal`) or copy operation (`Copy`) causes significant memory allocation overhead.
**Action:** Use a `sync.Pool` to reuse `[]byte` buffers to drastically reduce GC pressure and memory allocations.
