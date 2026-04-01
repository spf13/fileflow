## 2024-05-24 - TOCTOU Symlink Vulnerability in File Creation
**Vulnerability:** The `Copy` function uses `os.O_CREATE|os.O_TRUNC` after checking for file existence, creating a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could create a symlink between the check and creation to overwrite arbitrary files.
**Learning:** `Exists` check combined with `O_CREATE|O_TRUNC` is insufficient to prevent symlink attacks.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure atomic file creation and fail securely if the file was created in the interim.
