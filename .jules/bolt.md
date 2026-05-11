## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-11 - Pointer Conversion in sync.Pool
**Learning:** When using `sync.Pool` to reuse `[]byte` in Go, storing pointers (`*[]byte`) instead of the slices themselves avoids an extra heap allocation. This happens because converting a slice to `interface{}` inside `Put()` or `Get()` forces an allocation.
**Action:** Always store and retrieve pointers to slices in `sync.Pool` (`bufferPool.Get().(*[]byte)`) to maximize memory efficiency.
