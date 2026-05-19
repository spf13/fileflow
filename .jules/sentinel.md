## 2024-05-24 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability in `Copy` function when opening destination file.
**Learning:** Using `os.OpenFile` with `os.O_CREATE|os.O_TRUNC` directly allows an attacker to substitute a symlink at the destination path between checking if it exists and opening it, potentially overwriting arbitrary files.
**Prevention:** Use an atomic write pattern by creating a temporary file with `os.CreateTemp` in the destination directory, writing to it, securely applying permissions via `Chmod`, closing it, and then using `os.Rename` to atomically move it to the target destination.
