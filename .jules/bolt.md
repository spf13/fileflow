## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2026-06-04 - sync.Pool for buffer reuse
**Learning:** File operations (like `Equal` and `Copy`) that frequently allocate large `[]byte` buffers create significant memory allocations and GC pressure.
**Action:** Use a package-level `sync.Pool` to store and retrieve pointers to `[]byte` slices (`*[]byte`). Explicitly check capacity and reslice against any mutable global buffer size variables to safely reuse buffers and dramatically reduce allocations.
