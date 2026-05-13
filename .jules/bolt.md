## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-13 - Sync Pool buffer reuse
**Learning:** Using `sync.Pool` to reuse `[]byte` buffers for repeating file operations like `Equal` and `Copy` significantly reduces memory allocations and GC pressure. When pooling slices, using pointers (`*[]byte`) prevents an allocation during the conversion to `interface{}`.
**Action:** Implement `sync.Pool` for byte slice buffers, ensuring buffer length matches requirements using `cap` checking, and pool slice pointers instead of direct slices.
