## 2024-05-24 - TOCTOU Symlink Vulnerability in file copies
**Vulnerability:** The `Copy` function opens the destination file using `os.OpenFile(dst, ...)` after verifying its existence. This introduces a Time-of-Check to Time-of-Use (TOCTOU) vulnerability where an attacker could replace the destination with a symlink before `os.OpenFile` executes, causing the copy to follow the symlink and overwrite an unintended file.
**Learning:** Directly opening a file by name after checking its status is vulnerable to symlink attacks.
**Prevention:** Always use the atomic write pattern: write to a temporary file created via `os.CreateTemp` in the same directory, apply necessary permissions securely, and use `os.Rename` to atomically move it over the intended destination, preventing TOCTOU races.
