## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-02 - Optimize file copy and comparison operations with sync.Pool
**Learning:** In the fileflow package, allocating a new large buffer (e.g. 32KB) for every file comparison (`Equal`) or copy operation (`Copy`) causes significant memory allocation overhead.
**Action:** Use a `sync.Pool` to reuse `[]byte` buffers to drastically reduce GC pressure and memory allocations. When storing in `sync.Pool`, store pointers to `[]byte` (`*[]byte`) to prevent the slice-to-interface{} conversion from causing heap allocations. Also verify the slice capacity after retrieving it since `BufferSize` is a mutable global variable.
