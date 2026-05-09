## 2024-05-09 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** The `Copy` function in `fileflow.go` directly opens the destination path for writing, which is susceptible to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks if the destination changes between access check and file creation.
**Learning:** Simply appending `os.O_EXCL` breaks overwrite functionality. To securely fix TOCTOU without breaking features, an atomic write pattern is required.
**Prevention:** Use `os.CreateTemp` in the destination directory, apply original file permissions via `Chmod` before renaming, and atomically `os.Rename` the temporary file to the final destination. Explicitly clean up the temp file on error and handle `destFile = nil` to prevent double-closes in deferred cleanup.
