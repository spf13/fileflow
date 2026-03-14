## 2024-03-14 - TOCTOU Vulnerability in File Copy
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) race condition in `Copy` function due to using `os.O_TRUNC` instead of `os.O_EXCL` after checking if a file exists.
**Learning:** Checking for file existence and then opening it with `O_TRUNC` leaves a window where an attacker could create a symlink at the destination path, causing the copy operation to overwrite arbitrary files the user running the process has access to.
**Prevention:** Always use `os.O_EXCL` in combination with `os.O_CREATE` when the intent is to create a new file and ensure it doesn't already exist at the time of creation.
