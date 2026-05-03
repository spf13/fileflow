## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-03 - sync.Pool implementation for dynamically sized buffers
**Learning:** Using `sync.Pool` to reuse `[]byte` slices whose size depends on a global variable (like `BufferSize`) requires checking capacity upon retrieval. Converting a `[]byte` to an `interface{}` also causes heap allocation.
**Action:** Store `*[]byte` in `sync.Pool` and always verify `cap(*b)` upon retrieval, reallocating if the global required size has increased.
