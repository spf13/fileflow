## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-24 - Use sync.Pool to avoid allocations for temporary buffers
**Learning:** Storing `[]byte` in `sync.Pool` by storing their pointers (`*[]byte`) instead of the slices directly prevents native heap allocations when wrapping them inside `interface{}`.
**Action:** Always wrap `[]byte` in pointers when using `sync.Pool` to achieve zero-allocation reuse.
