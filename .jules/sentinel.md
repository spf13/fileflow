## 2024-04-03 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** The `Copy` function in `fileflow.go` used `os.O_CREATE|os.O_TRUNC` without `os.O_EXCL` after an existence check. This created a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink at the destination path between the check and the file creation, causing the copy operation to overwrite an arbitrary file.
**Learning:** Checking for file existence before creation is not sufficient to prevent overwriting files if symlinks can be introduced concurrently.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure the file is newly created and fails securely if it already exists.
