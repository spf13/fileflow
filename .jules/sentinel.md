## 2024-03-25 - TOCTOU Symlink Vulnerability in fileflow.go

**Vulnerability:** The `Copy` function was using `os.O_TRUNC` when creating new files, which is vulnerable to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks. An attacker could replace the destination file with a symlink after the existence check, causing the file content to be written to an arbitrary location.
**Learning:** `os.O_TRUNC` is not safe when copying files.
**Prevention:** To prevent TOCTOU vulnerabilities, always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
