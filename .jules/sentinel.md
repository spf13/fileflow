## 2024-04-17 - TOCTOU Symlink Vulnerability
**Vulnerability:** The code checks if a destination file exists, then creates it using `os.O_CREATE` but without `os.O_EXCL`. This creates a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink at the destination path between the check and the creation, causing the program to write to an arbitrary file.
**Learning:** In Go, when creating a file after an existence check, always use `os.O_EXCL` in conjunction with `os.O_CREATE` to ensure the file does not exist at the exact moment of creation.
**Prevention:** Add `|os.O_EXCL` to `os.OpenFile` flags when `os.O_CREATE` is used, to ensure atomic creation.
