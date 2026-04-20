## 2024-04-20 - TOCTOU Symlink Vulnerability in File Copies
**Vulnerability:** The codebase was directly opening target files using `os.OpenFile` with `O_TRUNC`, creating a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could replace the destination with a symlink between the check and open, causing arbitrary file truncation.
**Learning:** Using `O_EXCL` breaks expected overwrite functionality in `fileflow`. The only safe way to achieve atomic, symlink-safe overwrites is to use `os.CreateTemp` in the same directory, followed by `os.Rename`.
**Prevention:** Always use the atomic write pattern (`os.CreateTemp` + `os.Rename`) for file creation/overwrites to avoid TOCTOU issues.
