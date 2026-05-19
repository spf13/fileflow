## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-19 - sync.Pool Pointer Allocation

**Learning:** When using `sync.Pool` to manage `[]byte` slices in Go to avoid allocations, store and retrieve pointers to the slice (`*[]byte`) instead of the slice directly. Converting a `[]byte` to an `interface{}` when putting it into the pool natively causes an additional heap allocation. Furthermore, if the required slice capacity depends on a mutable global variable like `BufferSize`, always verify the slice capacity after retrieving it. Read the mutable global variable once into a local variable (e.g., `size := BufferSize`) before the check to prevent race conditions.
**Action:** Use `*[]byte` instead of `[]byte` in `sync.Pool`, and always read the global variable locally to check capacity of retrieved slices to prevent bounds issues.
