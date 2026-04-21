## 2026-04-21 - Remove bufio to enable zero-copy on file transfers
**Learning:** Wrapping an `*os.File` in a `bufio.Writer` before passing it to `io.Copy` in Go silently disables zero-copy system calls (like `copy_file_range` or `sendfile`), forcing all data through user-space allocations. This causes large memory overhead (from ~1KB/op to ~33KB/op) without providing meaningful throughput benefits.
**Action:** Do not use `bufio.Writer` on OS files when using `io.Copy`. Let `io.Copy` use `*os.File` directly so it can leverage fast-path system calls.
