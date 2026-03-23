## 2026-03-23 - [TOCTOU Vulnerability in File Creation]
**Vulnerability:** [Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability when using `os.O_CREATE|os.O_TRUNC` after checking for file existence using `Exists()`.]
**Learning:** [Using `os.O_TRUNC` after checking if a file exists allows an attacker to exploit the time gap between check and use by replacing the destination path with a symlink. This could allow an attacker to overwrite arbitrary files, as the `OpenFile` call would just truncate and follow the symlink.]
**Prevention:** [Always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check. This ensures that the file creation fails if the file (or a symlink) already exists.]
