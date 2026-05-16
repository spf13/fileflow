## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-16 - Use sync.Pool for byte slices and store pointers
**Learning:** Using `sync.Pool` to reuse large `[]byte` buffers for repeated operations (like `Equal` and `Copy`) significantly reduces memory allocations and GC pressure. Storing and retrieving pointers to the slice (`*[]byte`) instead of the slice directly avoids an additional heap allocation when converting to `interface{}`.
**Action:** Implement `sync.Pool` for buffer allocations, ensure pointers are stored, and verify slice capacity upon retrieval, dynamically reallocating if it doesn't match the required size.
