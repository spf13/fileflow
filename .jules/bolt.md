## 2024-04-18 - Avoid bufio wrapping with io.Copy for zero-copy transfers
**Learning:** Wrapping an `*os.File` in a `bufio.Writer` before passing it to `io.Copy` disables Go's ability to use zero-copy system calls (like `sendfile` or `copy_file_range`), leading to increased allocations and reduced performance.
**Action:** When copying data between two `*os.File` instances using `io.Copy`, avoid `bufio` to allow the runtime to use optimized system-level file copy operations.
