## 2024-04-15 - Prevent TOCTOU vulnerability in file copy
**Vulnerability:** The `Copy` function checks if the destination file exists and then creates it using `os.OpenFile` with `os.O_CREATE|os.O_TRUNC`. This creates a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink at the destination path between the check and creation.
**Learning:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to guarantee the file is newly created and mitigate symlink attacks.
**Prevention:** Use `os.O_EXCL` instead of `os.O_TRUNC` when creating new files after resolving names.
