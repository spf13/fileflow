## 2024-03-28 - TOCTOU vulnerability in file creation
**Vulnerability:** The `Copy` function uses `os.OpenFile` with `os.O_CREATE` but without `os.O_EXCL` after checking file existence. This creates a Time-of-Check to Time-of-Use (TOCTOU) race condition where an attacker could create a symlink at the destination path between the check and the file creation, leading to writing to arbitrary files.
**Learning:** Checking for file existence and then creating it without `O_EXCL` is inherently racy and unsafe when writing files.
**Prevention:** Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check to ensure the file is created atomically and avoid symlink attacks.
