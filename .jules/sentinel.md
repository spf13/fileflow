## 2025-04-25 - TOCTOU (Time-of-Check to Time-of-Use) Symlink Vulnerability in File Copies
**Vulnerability:** The application was directly opening destination files for writing (`os.OpenFile` with `os.O_TRUNC`), which is vulnerable to TOCTOU attacks if an attacker can swap the destination path with a symlink between existence checks and the `OpenFile` call.
**Learning:** Using `os.O_EXCL` prevents this but breaks expected overwrite functionality. Cross-device temp files cause `os.Rename` failures.
**Prevention:** Always use the atomic write pattern: create a temporary file *in the same directory* as the destination using `os.CreateTemp`, write the data, explicitly copy permissions via `Chmod`, and then perform an atomic `os.Rename()` over the final destination.
