## 2026-04-04 - TOCTOU Symlink Vulnerability in Copy
**Vulnerability:** `Copy` function had a Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability. It checked for file existence, and then later created the file with `os.OpenFile` using `os.O_CREATE|os.O_TRUNC` without `os.O_EXCL`. This allows an attacker to create a symlink at the destination path between the check and the open, potentially overwriting arbitrary files.
**Learning:** Checking for file existence before creation does not guarantee the file doesn't exist at the exact moment of creation due to race conditions.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure the operation fails if the file was created concurrently.
