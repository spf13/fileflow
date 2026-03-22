## 2024-05-24 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** The `Copy` function in `fileflow.go` used `os.O_TRUNC` when creating the destination file after checking for its existence, making it vulnerable to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks.
**Learning:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to prevent TOCTOU vulnerabilities.
**Prevention:** Use `os.O_EXCL` instead of `os.O_TRUNC` in `os.OpenFile` calls when creating files after existence checks.
