## 2024-05-24 - Fix TOCTOU vulnerability in file copy
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability in `Copy` function. The function checks for file existence and identity, but an attacker could replace the destination with a symlink before the `os.OpenFile` call, potentially overwriting arbitrary files.
**Learning:** `os.OpenFile` with `os.O_EXCL` breaks overwrite functionality. A robust atomic write pattern is needed.
**Prevention:** Always use atomic writes via temporary files (e.g., `os.CreateTemp` followed by `os.Rename`) when writing to files to prevent symlink attacks and race conditions. Ensure permissions are explicitly maintained via `Chmod` on the temporary file before renaming.
