## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-05 - Pool byte slices directly via pointers to avoid interface allocations
**Learning:** When putting `[]byte` into a `sync.Pool`, converting the slice natively to `interface{}` triggers a heap allocation. We noticed this when we originally attempted `bufferPool.Put(b)` vs `bufferPool.Put(&b)`.
**Action:** Store and retrieve pointers to slices (`*[]byte`) when using `sync.Pool` to avoid native `interface{}` conversion allocations.
