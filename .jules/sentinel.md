## 2024-05-24 - TOCTOU vulnerability in file copy
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) race condition in the `Copy` function file copying logic allows potential symlink attacks when creating destination files because `os.O_TRUNC` was used after an existence check.
**Learning:** Checking for file existence before opening/creating it without `os.O_EXCL` leaves a window where an attacker can create a symlink, tricking the application into overwriting unauthorized files.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
