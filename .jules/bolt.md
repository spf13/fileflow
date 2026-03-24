## 2024-11-06 - Remove bufio from file copy for zero-copy transfers
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`), degrading performance and increasing allocations. Direct use of `io.Copy` between two `*os.File` objects allows the OS to optimize the transfer.
**Action:** When copying data between two file descriptors in Go, prefer passing the `*os.File` instances directly to `io.Copy` rather than wrapping them in a buffered writer, unless there's a specific need for fine-grained writes.
