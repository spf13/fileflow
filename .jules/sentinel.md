## 2024-05-24 - TOCTOU Symlink Vulnerability in file copying
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) vulnerability exists in `fileflow.go` where `Exists` check is followed by `os.OpenFile(..., os.O_CREATE|os.O_TRUNC)`. An attacker could create a symlink in between, causing arbitrary file overwrite.
**Learning:** Checking for file existence before creating a file is prone to TOCTOU symlink attacks if the file creation doesn't enforce exclusivity.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
