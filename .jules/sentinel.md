## 2026-06-03 - Mitigate TOCTOU in file copy
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability in `Copy` function caused by creating a file and then writing to it without atomicity.
**Learning:** Using `os.OpenFile` with `os.O_EXCL` breaks overwrite functionality. An atomic write pattern (temp file + rename) is required, but it's important to set correct permissions (`Chmod`) before the rename.
**Prevention:** Use `os.CreateTemp` to create a temporary file in the destination directory, write data, sync, apply permissions, close, and finally rename atomically. Make sure to close and nil the file reference on success before `Rename` to prevent the defer block from removing the temporary file, but ensure it handles cleanup on error.
