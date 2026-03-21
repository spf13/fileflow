## 2026-03-21 - [Time-of-Check to Time-of-Use Symlink Vulnerability]
**Vulnerability:** [The `fileflow.go` package uses `Exists` to check for destination file presence and later opens the file with `os.O_CREATE` without `os.O_EXCL`. This creates a race condition where an attacker could create a symlink at the destination path between the check and the open operation, causing the application to write data to an unintended location.]
**Learning:** [In Go, when creating a file after checking its non-existence, `os.O_EXCL` must be used with `os.O_CREATE` in `os.OpenFile` to ensure atomicity and prevent TOCTOU symlink attacks.]
**Prevention:** [Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.]
