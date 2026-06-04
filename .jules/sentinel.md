
## 2026-06-04 - Prevent TOCTOU symlink vulnerabilities using atomic writes
**Vulnerability:** The `Copy` function used `os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)` directly, creating a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where `dst` could be replaced by a symlink after existence checks but before opening.
**Learning:** In Go, adding `os.O_EXCL` to prevent this breaks legitimate overwrite scenarios.
**Prevention:** Always use an atomic write pattern when creating/overwriting files: write to a temporary file via `os.CreateTemp(filepath.Dir(dst), ...)` in the same directory (to prevent cross-device rename issues), apply permissions with `Chmod`, write data, and finalize with an atomic `os.Rename`. Use a robust `defer` block to clean up the temporary file if errors occur.
