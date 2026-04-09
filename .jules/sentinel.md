## 2024-04-09 - [TOCTOU Vulnerability in File Copying]
**Vulnerability:** TOCTOU (Time-of-Check to Time-of-Use) race condition during file copying where destination file existence check is decoupled from creation.
**Learning:** In this codebase, to prevent TOCTOU symlink vulnerabilities, always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
**Prevention:** Always combine `os.O_CREATE` with `os.O_EXCL` when opening files to ensure atomic creation and avoid symlink attacks.
