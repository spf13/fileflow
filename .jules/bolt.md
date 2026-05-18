## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-18 - Use sync.Pool to reuse large []byte buffers
**Learning:** In operations that frequently allocate large `[]byte` buffers, such as `io.CopyBuffer` or file reading during comparison (`Equal`), significant overhead and GC pressure are incurred. Converting the slice to `interface{}` natively during pool insertion adds heap allocation overhead if not handled carefully (e.g. by using pointers). Also, buffer sizes from mutable global variables might change at runtime, so checking slice capacity before reuse is crucial.
**Action:** Use `sync.Pool` to store `*[]byte` pointers rather than `[]byte` values directly. Verify the capacity against the required size (e.g. `BufferSize`) after retrieval to handle variable-length buffers effectively and prevent panics.
