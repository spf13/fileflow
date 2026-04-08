## 2024-10-24 - Prevent TOCTOU Symlink Vulnerability in file creation
**Vulnerability:** The `fileflow.go` package created destination files using `os.O_CREATE` after checking for file existence. This left a Time-of-Check to Time-of-Use (TOCTOU) window where an attacker could create a symlink before the file was opened, potentially overwriting unintended files.
**Learning:** Using `Exists` followed by `os.OpenFile` without `os.O_EXCL` in Go is a common TOCTOU vulnerability when creating files.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure the file is newly created and not following a symlink.
