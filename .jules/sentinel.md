## 2024-05-24 - Prevent TOCTOU symlink vulnerability in Copy
**Vulnerability:** TOCTOU (Time-of-Check to Time-of-Use) symlink vulnerability when copying files, as `os.OpenFile` writes directly to the destination path.
**Learning:** Writing directly to the destination allows symlink attacks. Using `os.O_EXCL` breaks overwrite functionality. The atomic write pattern (temp file + rename) with `os.CreateTemp` is necessary.
**Prevention:** Always use the atomic write pattern (create temp file in target directory, set permissions, write data, sync, close, then atomically rename) when copying or moving files to prevent TOCTOU.
