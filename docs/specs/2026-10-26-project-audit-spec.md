# Living Specification: `fileflow` Hardening & Simplification (`2026-10-26`)

- **Source Audit**: `docs/audits/2026-10-26-project-audit.md` (`R-01` – `R-07`)
- **Status**: `READY FOR EXECUTION`

---

## 1. Core Invariants
1. **Strict No-Overwrite Guarantee**: `Move`, `Copy`, and `Rename` MUST NEVER overwrite an existing non-identical file unless `Flow.AllowOverwriteFallback` is explicitly opted into by the caller.
2. **Consistent Regular-File Source Gate**: `Move`, `Copy`, `Rename`, and `Replace` MUST all validate `os.Lstat(src)` and reject symlinks/non-regular files with `fs.ErrInvalid`.
3. **Uniform Error Wrapping & Durability**: All `Move`, `Copy`, `Rename`, and `Replace` failures wrap `Src` and `Dst` in `*ErrFailedMovingFile` (preserving `errors.Is` chains), and Unix publishes sync the parent directory descriptor.

---

## 2. EARS Requirements (`R-01` – `R-07`)
- **R-01**: **WHEN** conflict naming or publish retries run, **THE SYSTEM SHALL** use the shared `findAvailable` and `ensureDir` primitives.
- **R-02**: **WHEN** `Move`, `Copy`, or `Rename` is called on a symlink/non-regular source, **THE SYSTEM SHALL** return `fs.ErrInvalid`; **WHEN** hard links are unsupported on Unix and `AllowOverwriteFallback` is false, **THE SYSTEM SHALL** use `RENAME_NOREPLACE` (where available) or return an error rather than silently overwriting via `os.Rename`.
- **R-03**: **WHEN** a file is published via `Copy`, `Rename`, or `Replace` on Unix, **THE SYSTEM SHALL** sync the parent directory descriptor.
- **R-04**: **WHEN** `Copy` fails, **THE SYSTEM SHALL** return `*ErrFailedMovingFile{Src, Dst, Err}` with `%q`-quoted paths.
- **R-05**: **WHEN** `FindAvailableNameCont` parses trailing `-N` suffixes, **THE SYSTEM SHALL** use string scanning without `regexp`, with `go.mod` set to `go 1.24`.
- **R-06**: **WHEN** `go test -bench=.` runs, **THE SYSTEM SHALL** use `b.TempDir()`, `for b.Loop()`, and `os.Remove(final)`.
- **R-07**: **WHEN** `.jules/` or `*.tmp` files exist locally, **THE SYSTEM SHALL** ignore them via `.gitignore`.
