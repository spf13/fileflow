## 2024-05-24 - TOCTOU Symlink Vulnerability in File Creation
**Vulnerability:** The `Copy` function uses `Exists()` to check for destination file presence before calling `os.OpenFile` with `os.O_CREATE|os.O_TRUNC`. This creates a Time-of-Check to Time-of-Use (TOCTOU) race condition where an attacker could place a symlink at the destination path, causing arbitrary file overwrite.
**Learning:** Using an existence check before file creation without enforcing exclusivity during creation is insecure and vulnerable to race conditions.
**Prevention:** Always use `os.O_EXCL` alongside `os.O_CREATE` when creating files following an existence check to ensure atomic creation and prevent symlink attacks.
