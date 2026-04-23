## 2024-05-20 - TOCTOU Vulnerability in File Operations
**Vulnerability:** The application used `os.OpenFile` directly on the destination path which creates a Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability.
**Learning:** Writing directly to a target file can allow an attacker to swap the file for a symlink between checks and the write operation. Using `os.O_EXCL` breaks expected overwrite functionality.
**Prevention:** Use an atomic write pattern: create a temporary file in the destination directory using `os.CreateTemp(filepath.Dir(dst), ...)`, write data to it, set correct permissions via `Chmod`, and then perform an atomic `os.Rename()` over the intended destination.
