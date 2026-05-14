## 2024-05-14 - Fix TOCTOU Symlink Vulnerability in file Copy
**Vulnerability:** Time-of-Check to Time-of-Use (TOCTOU) vulnerability where file existence is checked, but a symlink could be created at `dst` before `os.OpenFile` is called, potentially overwriting unintended files.
**Learning:** Using `os.OpenFile` with `O_TRUNC` after a check without `O_EXCL` allows attackers to exploit the race condition.
**Prevention:** Use an atomic write pattern (write to a temporary file via `os.CreateTemp` and use `os.Rename`) to guarantee safe overwrites.
