## 2024-05-27 - [Fix TOCTOU Vulnerability in Copy Function]
**Vulnerability:** The `Copy` function was vulnerable to a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where `os.OpenFile` was used directly, which could allow an attacker to switch the file to a symlink before it was written.
**Learning:** Always use an atomic write pattern (temp file + rename) to prevent TOCTOU symlink vulnerabilities and cross-device rename failures. Make sure to securely set the permissions on the temporary file using `Chmod` before renaming it.
**Prevention:** Use `os.CreateTemp` to create a temporary file in the destination directory, apply original or intended permissions using `Chmod`, write the data, and then perform an atomic `os.Rename()` over the intended destination.
