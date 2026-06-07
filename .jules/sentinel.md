
## 2026-06-07 - [Atomic Write Pattern for TOCTOU]
**Vulnerability:** The application opened destination files using `os.OpenFile(..., os.O_CREATE|os.O_TRUNC, ...)`, leaving a Time-of-Check to Time-of-Use window where an attacker could replace the destination path with a symlink.
**Learning:** Using `os.O_EXCL` with `os.OpenFile` breaks legitimate overwrite behaviors. Additionally, defer cleanup patterns can leak resources if variables are not explicitly managed before returning errors.
**Prevention:** Always use an atomic write pattern when dealing with files by using `os.CreateTemp` in the destination directory, correctly managing deferred cleanup logic by setting references to nil upon success, copying over permissions with `Chmod`, and atomically switching with `os.Rename`.
