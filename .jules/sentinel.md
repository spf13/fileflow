## 2024-04-06 - Prevent TOCTOU symlink vulnerabilities in file creation
**Vulnerability:** The `Copy` function opens the destination file using `os.O_CREATE|os.O_TRUNC` after checking for its existence. This introduces a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink at the destination path between the check and the file creation, leading to arbitrary file overwrite.
**Learning:** In this codebase, to prevent TOCTOU symlink vulnerabilities, always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
**Prevention:** Use `os.O_EXCL` instead of `os.O_TRUNC` when creating files that are expected to be new, ensuring the operation fails if the file (or a symlink) was created concurrently.
