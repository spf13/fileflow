## 2024-05-30 - TOCTOU File Overwrite Vulnerability
**Vulnerability:** The `Copy` function opens a new destination file directly using `os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())`. This is susceptible to Time-Of-Check to Time-Of-Use (TOCTOU) symlink attacks where an attacker could replace `dst` with a symlink to another sensitive file between the check and the actual writing, causing an arbitrary file overwrite.
**Learning:** In Go, checking if a file exists and then writing to it directly creates a race condition.
**Prevention:** Use an atomic write pattern by writing to a temporary file (`os.CreateTemp`) in the same directory, setting its permissions via `Chmod`, and then atomically moving it over the target destination using `os.Rename`.
