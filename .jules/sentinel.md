## 2026-06-09 - Fix TOCTOU symlink vulnerability in file copying
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) vulnerability where `os.OpenFile` was directly overwriting the target destination path, which could be exploited if an attacker swaps the destination with a symlink before the file is written.
**Learning:** Using `os.O_TRUNC|os.O_CREATE` directly on destination files can lead to symlink attacks.
**Prevention:** Always use an atomic write pattern (create a temporary file in the destination directory, write the content, and then rename it over the destination atomically) to prevent TOCTOU symlink vulnerabilities and partial writes.
