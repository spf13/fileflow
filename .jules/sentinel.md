## 2024-03-20 - Prevent TOCTOU Symlink Vulnerability
**Vulnerability:** The application used os.O_CREATE without os.O_EXCL after checking file existence (Time-of-Check to Time-of-Use). This could allow an attacker to create a symlink between the check and creation, potentially overwriting arbitrary files.
**Learning:** In Go, checking if a file exists before creating it is inherently race-prone.
**Prevention:** Always use os.O_EXCL in conjunction with os.O_CREATE when opening or creating files after an existence check to ensure the file is created safely without overwriting an existing file or following a symlink.
