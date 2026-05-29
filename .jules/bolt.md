## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-29 - Use sync.Pool for FileFlow buffer allocations
**Learning:** In Go, converting a `[]byte` to an `interface{}` when putting it into a `sync.Pool` causes a heap allocation. Storing and retrieving pointers (`*[]byte`) instead avoids this overhead. Since `BufferSize` can be mutated globally, we must explicitly read it to a local variable to prevent TOCTOU race conditions when verifying capacity.
**Action:** Store `*[]byte` instead of `[]byte` in `sync.Pool` and read mutable globals to local variables before size validation.
