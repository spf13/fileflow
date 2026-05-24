## 2024-04-23 - Avoid bufio.Writer with io.Copy for os.File
**Learning:** Wrapping a standard file (`*os.File`) in `bufio.Writer` when using `io.Copy` disables zero-copy system calls like `sendfile` or `copy_file_range`. This degrades file copy performance.
**Action:** Use direct `io.Copy` with `*os.File` instances directly, instead of wrapping them in a buffered writer, to take advantage of OS-level zero-copy optimizations.
## 2024-05-24 - Document Optimizations for PR Compliance
**Learning:** PR validations and explicit constraints require any code optimization to include inline code comments explaining *why* the optimization exists, even if the commit message describes it well.
**Action:** Always verify that explicit, descriptive comments are injected directly alongside the implementation of an optimization (e.g., above a new sync.Pool) to pass the code review and strictly adhere to the prompt constraint: 'Add comments explaining the optimization'.
