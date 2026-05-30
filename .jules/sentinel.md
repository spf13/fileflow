## 2024-05-30 - TOCTOU Vulnerability in File Copy
**Vulnerability:** The `Copy` function used `os.OpenFile(dst, ...)` to write to a file after checking if it existed, creating a TOCTOU symlink race condition.
**Learning:** In Go, checking a file and then opening it natively is vulnerable. The atomic write pattern (temp file + `os.Rename`) is necessary to prevent this, and requires careful handling of permissions via `Chmod` and cleanup via `defer`.
**Prevention:** Always use atomic writes (`os.CreateTemp` followed by `os.Rename`) instead of direct file truncation when creating/overwriting files in directories that might be shared or accessible to other users.
