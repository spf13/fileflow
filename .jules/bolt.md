## 2024-04-06 - [Avoid bufio.Writer for io.Copy with os.File]
**Learning:** Wrapping `*os.File` in `bufio.Writer` before passing to `io.Copy` disables zero-copy system calls (like `sendfile` or `copy_file_range`) and degrades file copy performance.
**Action:** Directly pass `*os.File` to `io.Copy` to utilize zero-copy system calls when copying file contents.
