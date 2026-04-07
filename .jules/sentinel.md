## 2024-04-07 - Prevent TOCTOU symlink vulnerability in Copy function
**Vulnerability:** File copy operation was vulnerable to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks because `os.OpenFile` used `os.O_TRUNC` instead of `os.O_EXCL` after checking for file existence.
**Learning:** In Go, checking if a file exists before creating it is not thread-safe or secure against external filesystem changes. An attacker could create a symlink at the destination path between the check and the open call.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure atomic file creation.
