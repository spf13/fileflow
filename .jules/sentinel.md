## 2024-05-18 - TOCTOU Symlink Vulnerability in File Creation
**Vulnerability:** The application was using `os.O_TRUNC` alongside `os.O_CREATE` after checking for file existence. An attacker could exploit this Time-of-Check to Time-of-Use window by replacing the target with a symlink between the existence check and the open call, potentially overwriting sensitive files.
**Learning:** Checking if a file exists before creating it does not guarantee the file path is safe when the creation occurs, as the filesystem state can change concurrently.
**Prevention:** In this codebase, to prevent TOCTOU symlink vulnerabilities, always use `os.O_EXCL` in conjunction with `os.O_CREATE` when opening or creating files after an existence check.
