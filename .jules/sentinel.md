## 2024-05-24 - Fix TOCTOU vulnerability in file copying
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability existed in the `Copy` function. The code checked if a destination file existed, and if not, proceeded to create it with `os.O_CREATE|os.O_TRUNC`. An attacker could create a symlink at the destination path between the check and the creation, causing the program to overwrite an arbitrary file the user has permissions to.
**Learning:** Checking for file existence before creation without `os.O_EXCL` is inherently racy and vulnerable to TOCTOU attacks.
**Prevention:** In this codebase, to prevent TOCTOU symlink vulnerabilities, always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
