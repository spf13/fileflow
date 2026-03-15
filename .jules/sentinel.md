## 2024-03-15 - TOCTOU Vulnerability in File Creation
**Vulnerability:** The `Copy` function was using `os.O_TRUNC` to create files after an existence check, which created a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where a symlink could be created in between the check and creation.
**Learning:** Checking if a file exists and then conditionally creating it leaves a window where the state of the filesystem can change. Always use atomic operations for file creation when possible.
**Prevention:** Use `os.O_EXCL` in conjunction with `os.O_CREATE` when you expect a file not to exist. This ensures that the open operation will fail atomically if the file was created concurrently.
