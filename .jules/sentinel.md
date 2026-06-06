
## 2026-06-06 - [TOCTOU Vulnerability in File Copy]
**Vulnerability:** The `Copy` function in `fileflow.go` used `os.OpenFile` directly on the destination path with `O_CREATE|O_TRUNC`.
**Learning:** If an attacker can predict the destination filename and create a symlink to a sensitive file before the copy operation executes, the program will unwittingly follow the symlink and overwrite the sensitive file.
**Prevention:** Always use the atomic write pattern when creating or modifying files: create a secure temporary file in the destination directory using `os.CreateTemp`, write data to it, apply required permissions, and then atomically `os.Rename` it over the intended destination.
