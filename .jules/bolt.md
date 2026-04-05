## 2024-04-05 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping standard files (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls (`sendfile`, `copy_file_range`) and degrades file copy performance.
**Action:** Always pass `*os.File` directly to `io.Copy` when copying files to utilize zero-copy system calls.
