## 2024-05-23 - Prevent TOCTOU symlink vulnerabilities
**Vulnerability:** os.OpenFile with O_CREATE|O_TRUNC directly to the target path is vulnerable to Time-of-Check to Time-of-Use (TOCTOU) symlink attacks.
**Learning:** To prevent TOCTOU vulnerabilities during file writes, use an atomic write pattern by creating a temporary file and renaming it.
**Prevention:** Use os.CreateTemp and os.Rename, maintaining the original file permissions with Chmod. Ensure correct defer cleanup for temporary files to avoid leaks on error.
