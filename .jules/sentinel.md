## 2024-05-18 - TOCTOU Symlink Vulnerability in file creation
**Vulnerability:** `os.OpenFile` was used with `os.O_CREATE|os.O_TRUNC` directly on the target destination, leaving a window for Time-of-Check to Time-of-Use (TOCTOU) symlink attacks.
**Learning:** Adding `os.O_EXCL` breaks overwrite functionality. Instead, an atomic write pattern with `os.CreateTemp` and `os.Rename` must be used to safely write and overwrite files.
**Prevention:** Always write to a temporary file in the destination directory and then atomically rename it to the target destination.
