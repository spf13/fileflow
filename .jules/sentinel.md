## 2026-06-05 - [Sentinel]
**Vulnerability:** TOCTOU (Time-of-Check to Time-of-Use) symlink vulnerability in `os.OpenFile` during file write/copy operations.
**Learning:** Directly using `os.OpenFile` with just `os.O_EXCL` breaks overwrite functionality. To prevent TOCTOU securely while preserving overwrites, an atomic write pattern is required.
**Prevention:** Create a temporary file in the destination directory (`os.CreateTemp(filepath.Dir(dst), ...)`), write the data, securely apply intended file permissions using `Chmod` (`destFile.Chmod(sourceInfo.Mode())`), and perform an atomic `os.Rename()` over the destination.
