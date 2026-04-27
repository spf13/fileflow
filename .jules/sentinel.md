## 2024-04-27 - TOCTOU Vulnerability in File Copy
**Vulnerability:** The `Copy` function opens the destination file with `os.O_RDWR|os.O_CREATE|os.O_TRUNC` directly, creating a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where the file can be swapped for a symlink between checking its existence and writing to it.
**Learning:** Always use the atomic write pattern (temp file + rename) to prevent TOCTOU issues. When using the atomic write pattern, ensure to properly set the temporary file permissions to match the original using `Chmod` before renaming.
**Prevention:** Use `os.CreateTemp` in the same directory, write data, apply permissions, close, and then use `os.Rename` for atomic file creation.
