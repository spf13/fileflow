## 2024-05-11 - TOCTOU Symlink Vulnerability in File Copies
**Vulnerability:** The `Copy` function opens the destination file directly with `os.OpenFile` and `os.O_TRUNC`, creating a Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability.
**Learning:** To avoid TOCTOU vulnerabilities and ensure atomic writes, we must use an atomic write pattern (temp file + rename) and apply permissions using `Chmod` before the rename. Setting file to `nil` in defer when closing fails prevents double-close leaks.
**Prevention:** Always use `os.CreateTemp` in the destination directory followed by `os.Rename` for file modifications, and securely copy file permissions via `Chmod` to the temp file.
