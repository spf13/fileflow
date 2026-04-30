## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.

## 2024-05-01 - sync.Pool allocation reduction
**Learning:** In Go, converting a `[]byte` to an `interface{}` when putting it into a `sync.Pool` natively causes an additional heap allocation. Furthermore, repeated allocation of large buffers (e.g. 32KB) creates significant GC overhead.
**Action:** Use `sync.Pool` to reuse large `[]byte` slices, and explicitly store/retrieve pointers to the slices (`*[]byte`) to prevent the interface conversion allocation.
