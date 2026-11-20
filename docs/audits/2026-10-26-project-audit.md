# Project Audit: `fileflow` (coupled with `pathologize`)

- **Date**: 2026-10-26
- **Target**: `/share/steve/code/fileflow` (`12 files`, `2,524 LOC`) + `/share/steve/code/pathologize` (`5 files`, `930 LOC`)
- **Surfaces Audited**: `Go Library / SDK (Move, Copy, Rename, Replace, Exists, Equal, FindAvailableName*)`
- **Lenses Run**: `PM, DX/SDK Design, Architecture, Engineering` + `Adversarial Skeptic`
- **Verification Summary**: `20 raw findings → 2 refuted / 11 merged by skeptic → 7 confirmed recommendations`
- **Estimated Simplification Leverage**: `−75 LOC duplicated retry/naming loops, −1 regexp init dependency, 3 filesystem safety/durability gaps closed (FAT silent overwrite, Lstat symlink parity, directory fsync)`

---

## 1. Executive Diagnosis & Goal Alignment

- **Stated Goal**: Safe, race-resistant file `Move`, `Copy`, `Rename`, and atomic `Replace` for Go (`spf13/fileflow`), guaranteeing no accidental overwrite of non-identical files, seamless `EXDEV` cross-device fallback, and conflict naming strategies—paired with `spf13/pathologize` for untrusted path sanitization.
- **Codebase Reality**: A compact, zero-global-state Go library (`fileflow.go` 503 LOC, `replace.go` ~120 LOC, `1,920 LOC` of tests).
- **Core Alignment Gaps**:
  1. **Silent No-Overwrite Violation on Non-Hardlink Unix Volumes (`FATAL`)**: When `os.Link` fails with `EPERM`/`ENOTSUP`/`ENOSYS` (FAT/exFAT/SMB/CIFS mounts), `fileflow_notwindows.go:28-36` silently falls back to `os.Rename(src, dst)`—which overwrites any concurrently created destination file and makes the retry loop's `errors.Is(err, fs.ErrExist)` check dead code.
  2. **Inconsistent Symlink Gate & Missing Parent-Directory `fsync`**: `Replace` (`replace.go:34`) checks `os.Lstat(src)` and rejects symlinks/non-regular files, but `Move`, `Copy`, and `Rename` use `os.Open`/`os.Stat` and follow symlinks. Meanwhile, `copyToTemp` (`fileflow.go:489`) calls `tmp.Sync()` without `Sync()`ing the parent directory descriptor after rename.
  3. **Inconsistent Error Wrapping in `Copy` & Dishonest `go 1.19` Floor**: `Rename` and `Replace` wrap failures in `*ErrFailedMovingFile{Src, Dst, Err}`, while `Copy` returns bare `fmt.Errorf` without `Src`/`Dst` fields. Meanwhile, `go.mod` declares `go 1.19` even though its required `golang.org/x/sys v0.30.0` dependency requires `go 1.23.0`.
- **Strongest Existing Assets**: Clean stateless `Flow` struct design (replacing legacy package globals), byte-identical deduplication before conflict suffixing, and high-concurrency race tests (`TestCopyNeverOverwritesUnderRace`).

---

## 2. Prioritized Recommendations Matrix

Ordered by the **Leverage Hierarchy** (`DELETE/COLLAPSE` → `FIX_FLOW` → `SIMPLIFY/ARCHITECT` → `FEATURE_GAP`):

| ID | Category | Recommendation | Lenses | Severity | Effort | Net Delta |
|---|---|---|---|---|---|---|
| **R-01** | `COLLAPSE` | Collapse 4 `FindAvailableName*` loops into 1 helper & unify `Rename`/`Copy` retry loop | ENG | `MEDIUM` | `S` | `−75 LOC` |
| **R-02** | `FIX_FLOW` | Gate Unix `os.Rename` overwrite fallback behind `AllowOverwriteFallback` (or `RENAME_NOREPLACE`) & enforce `Lstat` regular-file check across `Move`/`Copy`/`Rename` | PM, ARCH | `FATAL` | `S` | `+30 LOC` |
| **R-03** | `FIX_FLOW` | Add parent-directory `fsync` (`syncDir`) after publish in `Copy`, `Rename`, and `Replace` on Unix | ARCH | `HIGH` | `S` | `+15 LOC` |
| **R-04** | `FIX_FLOW` | Wrap `Copy` errors in `*ErrFailedMovingFile`, quote paths with `%q`, & document `(dst, *ErrFailedRemovingOriginal)` committed-state contract | UX, ARCH, ENG | `HIGH` | `S` | `+25 LOC` |
| **R-05** | `SIMPLIFY` | Align `go.mod` floor (`go 1.24`), replace `contSuffixPattern` regex with string scan, & sync `godoc` with `readme.md` | PM, UX, ARCH, ENG | `MEDIUM` | `S` | `−15 LOC, −1 regexp` |
| **R-06** | `FIX_FLOW` | Fix `fileflow_bench_test.go` (`b.TempDir()`, `b.Loop()`, `b.SetBytes`, `os.Remove(final)`) & add `BenchmarkEqual` / `BenchmarkReplace` | ENG | `MEDIUM` | `S` | `+25 LOC` |
| **R-07** | `DELETE` | Add `.jules/` and `*.tmp` ignore rules to `.gitignore` in `fileflow` and `pathologize` | PM | `LOW` | `S` | `+2 lines` |

---

## 3. Detailed Recommendations (Confirmed)

### [R-01] Collapse 4 `FindAvailableName*` Loops & Unify `Rename`/`Copy` Retry Loop
- **Category**: `COLLAPSE` | **Lenses**: `ENG` | **Severity**: `MEDIUM` | **Effort**: `S`
- **Evidence**: `fileflow.go:235-312` repeats extension splitting and `for i := 1; i <= maxIncrementAttempts; i++` across `Inc`, `Cont`, `Auto`, and `TS`; `fileflow.go:160-196` (`Rename`) and `fileflow.go:370-410` (`Copy`) duplicate the `MkdirAll` + `resolveDst` + retry ceremony.
- **Recommended Change**: Extract internal `findAvailable(baseName, nextFn)` and `ensureDir(dst)` helpers while keeping all exported function signatures unchanged.
- **Verification Criteria**: `go test -race ./...` passes; `fileflow.go` shrinks by ~75 LOC.

### [R-02] Fix Silent Unix Overwrite Fallback & Enforce `Lstat` Source Validation
- **Category**: `FIX_FLOW` | **Lenses**: `PM, ARCH` | **Severity**: `FATAL` | **Effort**: `S`
- **Evidence**:
  - `fileflow_notwindows.go:28-36` falls back from `os.Link` to overwriting `os.Rename` on `EPERM`/`ENOTSUP`/`ENOSYS`/`EMLINK`.
  - `replace.go:34-42` checks `os.Lstat(src)` + `!Mode().IsRegular()`, whereas `Move`, `Copy`, and `Rename` follow symlinks via `os.Open`/`os.Stat`.
- **Recommended Change**: Refuse overwriting `os.Rename` unless `Flow.AllowOverwriteFallback` is explicitly set (or use `unix.Renameat2(..., RENAME_NOREPLACE)` on Linux), and apply the `os.Lstat(src)` regular-file check to `Move`, `Copy`, and `Rename`.
- **Verification Criteria**: Unit tests verify symlink sources return `fs.ErrInvalid` and non-hardlink fallback does not silently clobber existing files.

### [R-03] Sync Parent Directory After Publish (`syncDir`)
- **Category**: `FIX_FLOW` | **Lenses**: `ARCH` | **Severity**: `HIGH` | **Effort**: `S`
- **Evidence**: `fileflow.go:489` calls `tmp.Sync()` on file contents but never `Sync()`s the destination directory descriptor after `renameNoReplace` / `renameReplace`.
- **Recommended Change**: Add a Unix `syncDir(filepath.Dir(dst))` helper called after successful publish in `Copy`, `Rename`, and `Replace`.
- **Verification Criteria**: `go test ./...` passes on Unix and Windows.

### [R-04] Unify `*ErrFailedMovingFile` Wrapping in `Copy` & Document `*ErrFailedRemovingOriginal`
- **Category**: `FIX_FLOW` | **Lenses**: `UX, ARCH, ENG` | **Severity**: `HIGH` | **Effort**: `S`
- **Evidence**: `fileflow.go:410-500` (`Copy` / `copyToTemp`) returns bare `fmt.Errorf` without `Src`/`Dst` wrapper, unlike `Rename` (`fileflow.go:244`) and `Replace` (`replace.go:36,50,60`). `fileflow.go:109,125` formats paths with `%v` instead of `%q`.
- **Recommended Change**: Wrap `Copy` errors in `*ErrFailedMovingFile{Src: src, Dst: dst, Err: err}`, use `%q` in `Error()`, and document on `ErrFailedRemovingOriginal` that a non-empty returned path means the destination is already committed.
- **Verification Criteria**: `errors.As(err, &moveErr)` succeeds on `Copy`, `Move`, `Rename`, and `Replace` failures while preserving `errors.Is` unwrap chains.

### [R-05] Align `go.mod` Floor (`go 1.24`), Remove `regexp` Init, & Sync Godoc with `readme.md`
- **Category**: `SIMPLIFY` | **Lenses**: `PM, UX, ARCH, ENG` | **Severity**: `MEDIUM` | **Effort**: `S`
- **Evidence**: `go.mod:3` states `go 1.19` while `golang.org/x/sys v0.30.0` requires `go 1.23.0`; `fileflow.go:27` compiles `contSuffixPattern = regexp.MustCompile(`-(\d+)$`)` at package init.
- **Recommended Change**: Update `go.mod` to `go 1.24`, replace `contSuffixPattern` with a string suffix scan, and update `Flow`, `Exists`, and `Equal` godoc comments to match `readme.md`.
- **Verification Criteria**: `go vet ./...` and `go test ./...` pass with zero `regexp` import in `fileflow.go`.

### [R-06] Modernize `fileflow_bench_test.go` & Add `BenchmarkEqual` / `BenchmarkReplace`
- **Category**: `FIX_FLOW` | **Lenses**: `ENG` | **Severity**: `MEDIUM` | **Effort**: `S`
- **Evidence**: `fileflow_bench_test.go:11-28` uses `os.MkdirTemp`, legacy `for i := 0; i < b.N; i++`, and `os.Remove(dst)` instead of `os.Remove(final)`.
- **Recommended Change**: Use `b.TempDir()`, `b.ReportAllocs()`, `b.SetBytes()`, `for b.Loop()`, and `os.Remove(final)`, plus add `BenchmarkEqual` and `BenchmarkReplace`.
- **Verification Criteria**: `go test -bench=. -benchtime=10ms ./...` passes cleanly.

### [R-07] Guard `.gitignore` Against Bot Scratchpads
- **Category**: `DELETE` | **Lenses**: `PM` | **Severity**: `LOW` | **Effort**: `S`
- **Evidence**: Git history contains `.jules/bolt.md` (`fileflow` and `pathologize`) and `commit_msg.tmp` (`pathologize`).
- **Recommended Change**: Add `.jules/` and `*.tmp` to `.gitignore`.
- **Verification Criteria**: `.gitignore` blocks `.jules/` and `*.tmp`.

---

## 4. Rejected / Deferred Ideas (Skeptic Log)
- **Rename exported `ErrFailedMovingFile` struct to `MoveError` (`ENG-02`)**: *Refuted* — breaking change to exported types used by downstream `errors.As` callers and `skill://fileflow-pathologize`.
- **Add `pathologize` as a required dependency of `fileflow` (`PM-04`)**: *Refuted* — keeping `fileflow` focused on OS file movement and `pathologize` focused on lexical sanitization preserves minimal dependencies for consumers who only need one.
