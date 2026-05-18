## 2024-05-18 - TOCTOU Symlink Vulnerability in File Copy
**Vulnerability:** A Time-of-Check to Time-of-Use (TOCTOU) symlink vulnerability during file copies could occur, as the code checked if a destination existed and then opened it without O_EXCL.
**Learning:** To allow overwrites while preventing TOCTOU races, you must use an atomic write pattern (temp file + rename). Adding O_EXCL breaks standard overwrite behavior. You must also explicitly set permissions on the temp file before renaming.
**Prevention:** Always use atomic writes (CreateTemp in the destination directory, Write, Chmod, Rename) rather than OpenFile directly on the destination for file copying where overwrites might occur.
