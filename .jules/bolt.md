## 2024-04-10 - Avoid bufio with io.Copy for standard files
**Learning:** Wrapping `*os.File` in `bufio.Writer` disables zero-copy optimizations (`sendfile`, `copy_file_range`) when using `io.Copy`, significantly increasing memory allocations and CPU time.
**Action:** Directly pass `*os.File` to `io.Copy` to allow standard library zero-copy optimizations.
