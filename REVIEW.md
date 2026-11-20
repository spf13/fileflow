# Review guidelines: fileflow

Small, dependency-light Go 1.24 package (`golang.org/x/sys` only) that moves, copies, renames, and explicitly replaces files, including across filesystems. Platform code is split across `fileflow_windows.go`/`fileflow_notwindows.go` and `replace_windows.go`/`replace_notwindows.go`.

## Flag (medium or higher)
- The core safety guarantee in `readme.md`: `Move`, `Copy`, and `Rename` must never overwrite a non-identical file, including one created by a concurrent process between conflict resolution and the rename. Anything that introduces a delete-then-move window, drops the no-replace primitive, or lets `AllowOverwriteFallback` default to true.
- `ErrSameFile` detection: `sameFileReplace` (`replace.go`) must keep using `os.Lstat` on `dst`, so `Replace` replaces a `dst` symlink pointing at `src` instead of returning `ErrSameFile`. `sameFile` (`fileflow.go`) uses `os.Stat` on purpose for `Move`/`Copy`/`Rename` — do not flag it. `Replace` does its own inline `os.Lstat(src)` + `IsRegular` gate; `validateRegularFile` (`Lstat`) gates the source for `Move`/`Copy`/`Rename` only.
- Cross-device correctness: staging temp files in the destination directory, sync before publish, `syncDir` on the parent after publishing, source-mode preservation, and cleanup of the temp file on every error path.
- The documented `ErrFailedRemovingOriginal` contract (destination committed, source remains) and `*ErrFailedMovingFile` wrapping must stay accurate — a change that inverts or drops the wrap is a defect, as is `%q`-quoting regressions in `Error()` strings.
- Commit-message/diff gaps: when the message claims a rule (R-03 fsync, R-04 quoting/godoc, R-07 ignores), verify the diff actually contains it (job 1771 claimed `%q` quoting and the godoc update that were absent).
- If the commit references a Kata issue (`kata#<id>`), check the change against that issue's acceptance criteria and report gaps.

## Don't flag
- The `best-effort` no-overwrite caveat on hardlink-less Unix filesystems — documented, not a defect (jobs 1770, 1772 passed on this).
- Benchmark style (`b.Loop`, `b.TempDir`, `ReportAllocs`) and test-helper placement.
- Package-level naming/godoc polish already covered by the R-04 sweep, and duplicated `ErrFailedMovingFile` wording across Rename/Copy/Replace now that `publishWithRetry` is shared.
- Style, naming, comment or doc wording, and anything gofmt/linters already enforce.

## Severity
- high: reachable bug, regression, data loss (silent overwrite, lost source, partial destination), or security hole; contract or acceptance-criteria gap.
- medium: real defect with limited reach, or a missing test for new non-trivial logic (e.g. new error wrapping with no coverage).
- low: everything else (not reported; review_min_severity is medium).
