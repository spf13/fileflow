## 2024-04-26 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** `os.OpenFile` used directly on destination paths is vulnerable to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks.
**Learning:** Using `os.O_EXCL` breaks intended overwrite functionality. An atomic write pattern (temp file + rename) must be used.
**Prevention:** Create a temporary file via `os.CreateTemp(filepath.Dir(dst), ...)`, write contents, apply original permissions using `Chmod`, and atomically `os.Rename` it to the destination.
