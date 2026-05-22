## 2024-05-22 - Prevent TOCTOU Symlink Vulnerability in file writes
**Vulnerability:** `os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, ...)` directly opens the destination path. If an attacker replaces the destination with a symlink before this call, the program can overwrite arbitrary files (Time-of-Check to Time-of-Use).
**Learning:** Always use atomic writes by writing to a temporary file (`os.CreateTemp` in the same directory to avoid cross-device rename issues) and replacing the destination with `os.Rename`.
**Prevention:** Use atomic temporary files and explicit `destFile.Chmod` to preserve original file permissions. Manage defer cleanups tightly to prevent resource leaks when manually calling `.Close()`.
