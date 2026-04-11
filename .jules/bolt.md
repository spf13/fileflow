## 2024-04-11 - [Optimize File Copy]
**Learning:** Wrapping `*os.File` in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`), degrading file copy performance significantly.
**Action:** Avoid wrapping standard files (`*os.File`) in `bufio.Writer` when copying files.
