## 2024-06-25 - Fix TOCTOU vulnerability in Copy
**Vulnerability:** Creating a file and then modifying it in place makes the operation susceptible to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks.
**Learning:** Using `os.O_EXCL` breaks overwrite functionality. To securely overwrite while preventing TOCTOU attacks, we must use an atomic write pattern (temp file + `os.Rename`).
**Prevention:** Create a temporary file in the destination directory using `os.CreateTemp(filepath.Dir(dst), ...)` (to prevent cross-device rename failures), write the data, explicitly apply original file permissions to the temp file securely via `Chmod`, and then perform an atomic `os.Rename()` over the intended destination. Ensure temporary files are explicitly removed on cleanup errors.
