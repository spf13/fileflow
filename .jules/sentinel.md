## 2026-03-18 - TOCTOU Symlink Vulnerabilities
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) vulnerability exists when opening/creating files with `os.OpenFile` using `os.O_TRUNC` after a separate existence check. An attacker could replace the intended target with a symlink between the check and the open call, potentially overwriting arbitrary files.
**Learning:** `os.O_TRUNC` does not protect against symlink attacks if the file was replaced after an existence check. The existence check and the file creation must be atomic.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure the operation is atomic and fails if the file already exists (or is a symlink created in the interim).
