## 2024-05-15 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** The `Copy` function checked if a file existed, then used `os.OpenFile` with `os.O_TRUNC` to create or overwrite it. An attacker could swap the destination with a symlink between the check and the open, causing `fileflow` to overwrite an arbitrary file.
**Learning:** Checking for file existence before creation without enforcing mutual exclusion is a classic Time-of-Check to Time-of-Use (TOCTOU) vulnerability.
**Prevention:** When creating new files, always use `os.O_EXCL` in conjunction with `os.O_CREATE` to ensure that the file is created safely without following symlinks or overwriting newly placed files.
