## 2024-05-24 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) vulnerability exists in the `Copy` function where `os.OpenFile` uses `os.O_TRUNC` instead of `os.O_EXCL` after checking file existence.
**Learning:** Using `os.O_TRUNC` after an existence check allows a race condition where an attacker could create a symlink between the check and the open, causing the application to overwrite an arbitrary file.
**Prevention:** Always use `os.O_EXCL` with `os.O_CREATE` when opening or creating files after an existence check to ensure the file is newly created and not a symlink.
