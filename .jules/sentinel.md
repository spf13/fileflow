## 2024-04-14 - Prevent TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** The `Copy` function opens the destination file using `os.O_CREATE` without `os.O_EXCL` after checking for its existence. This creates a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink at the destination path between the check and the open, potentially overwriting arbitrary files.
**Learning:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
**Prevention:** Enforce the use of `os.O_EXCL` for all `os.OpenFile` calls that follow an existence check to ensure atomic file creation.
