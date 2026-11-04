# Ratchet report — fileflow, 2026-11-04

Branch `claude/cool-faraday-t8kd94`, base `4be9157`. 3 sweeps plus a bookend pass. 13 changes kept, 0 reverted. A separate judge reviewed every change, and a final reality gate ran on a fresh clone.

## Done signal: 7 of 8 met (NOT fully met)

| Item | Met | Evidence |
|---|---|---|
| Directory-destination bug fixed, test failed first | yes | `TestOntoDirectoryFailsFast` fails on 4be9157 and passes at HEAD |
| Dotfile bug fixed, test failed first | yes | `TestFindAvailableNameDotfiles` fails on 4be9157 and passes at HEAD |
| EPERM-overwrite bug fixed, test failed first | **no** | See "Correction" below |
| `Equal` at least 5× faster on 100 MB | yes | ~890 ms → ~43 ms (~20×), measured by the final gate |
| gofmt clean | yes | `gofmt -l .` is empty |
| errcheck in production code addressed | yes | golangci-lint 0 issues (baseline 11 shown, 12 real) |
| Suite green on go1.24 and go1.19 | yes | `-race -count=3` ok; `GOTOOLCHAIN=go1.19.13` ok |
| No exported API change | yes | `go doc -all` declarations identical on linux, windows, darwin and freebsd; go.mod unchanged |

**Correction to my Phase C claim.** I said link+remove "silently overwrites" when `fs.protected_hardlinks` makes `link` fail with EPERM. That is wrong when the destination already exists: the kernel returns EEXIST before it checks link permission, so the fallback never runs. The overwrite can only happen in a race, when the destination appears between the failed `link` and the fallback `os.Rename`. I could not write a deterministic test for that. Two things remain:
- On Linux and macOS the race is gone, because `renameat2` and `renamex_np` bypass `link`.
- On FreeBSD and other Unix systems, and wherever the atomic call returns ENOSYS or EINVAL, the EPERM → `os.Rename` race remains, as before.

The test that does fail first, `source_not_removable`, covers a different real bug: a failed rename left `dst` behind as a hard link to another user's file.

## Metrics

| Metric | Baseline (4be9157) | Final |
|---|---|---|
| Build / vet / tests | pass | pass |
| Tests + subtests run | 93 | 139 |
| Coverage | 83.4% | 84.9% |
| golangci-lint | 11 (12 real) | 0 |
| gofmt -l | 2 files | 0 |
| Golden smoke | 8/14 | 14/14 |
| `Equal`, 100 MB identical files | ~865–890 ms | ~38–43 ms |
| BenchmarkEqual, 10 MiB | ~105 MB/s | ~3.1–4.4 GB/s |
| BenchmarkCopy, 10 MB | ~22.9 ms | ~23.4 ms (noise) |
| Cold build | ~8.3 s | ~7.8 s |
| Production Go lines | 714 | 825 (+3 small rename files and doc comments) |
| Module requirements | x/sys v0.30.0 | unchanged |

## Kept changes

| sha | Change | Judge score |
|---|---|---|
| 47912f9 | CRLF → LF in fileflow.go and readme.md; gofmt replace.go (whitespace only) | — |
| 8b045a2 | Any existing entry is an occupied name. A directory destination (or a symlink to one) fails at once with `*ErrFailedMovingFile` matching `fs.ErrExist`, instead of 100 retries and `ErrMaxAttemptsReached`. A dangling symlink gets a suffix. Strategies skip directories. | 55→72 |
| fdd91fc | Check the source first: a missing source no longer creates destination directories, and is reported as `ErrNotExist`, not as a destination problem. | 55→66 |
| 38a057d | A leading dot belongs to the name: `.bashrc` → `.bashrc-1`, not `-1.bashrc`. A shared `splitExt` replaces 4 copies of the split logic. | 55→85 |
| 03d1959 | Atomic no-replace rename: `renameat2(RENAME_NOREPLACE)` on Linux, `renamex_np(RENAME_EXCL)` on macOS, with link+remove as fallback. | 60→85 |
| 7f6da45 | `Equal` compares 64 KiB chunks instead of single bytes. Adds BenchmarkEqual and chunk-boundary tests. | 20→90 |
| 84af3c7 | Remove `Move`'s duplicate same-file check. | 72→76 |
| 8bde877 | Document directory destinations, occupied names, dotfiles and the rename primitives. | 70→84 |
| 73c1085 | Make ignored cleanup errors explicit (`_ =`); the benchmark uses `b.TempDir`. | 55→80 |
| 53231c8 | CI: one `gofmt -l` step (ubuntu/stable). This is the only CI change, as BRIEF allows. | 50→85 |
| 8579d37 | Test the link+remove fallback directly (`renameLink`); the unprivileged helper test skips when it can't run. | 40→72 |
| 20e721f | Fix stale README lines; the custom-strategy example now uses `os.Lstat`; list Replace's errors. | 72→82 |
| aed4256 | Correct the `Exists` and `ErrMaxAttemptsReached` descriptions. | KEEP (final gate) |
| 5127615 | `.ratchet/check.sh` now fails on gofmt, lint and smoke regressions (tooling only). | verified by tamper test |

## Reverted attempts

None. One self-correction before committing: the first unprivileged-rename test passed on the old code, so I rewrote it around the sticky-directory case, which reproduces a real failure.

## Behaviour changes callers can see (all sanctioned by BRIEF)
- Move, Copy or Rename onto an existing directory returns `*ErrFailedMovingFile` wrapping `fs.ErrExist`. It used to return `ErrMaxAttemptsReached` after 100 retries.
- A dangling symlink or special file at the destination gets a conflict suffix. It used to fail with `ErrMaxAttemptsReached`.
- Dotfile conflict names changed: `.bashrc-1`, `.config-1.json`, and `.env-3` → `.env-4` under Cont and Auto.
- A missing source fails before any directory is created. Rename and Move return `*ErrFailedMovingFile`; Copy returns its usual "opening source file" error.
- When the source can't be removed, the rename fails cleanly with EPERM on Linux and macOS. It used to return a path plus `ErrFailedRemovingOriginal`, with a hard-link alias left behind.

## Remaining backlog (ranked)

| # | Item | Effort | Notes |
|---|---|---|---|
| 1 | The EPERM → `os.Rename` fallback race on non-Linux/macOS Unix systems | M | FreeBSD has `renameat` but no no-replace flag. Options: keep it as best-effort and document it, or refuse instead of falling back. This changes behaviour on those platforms, so it's your decision. |
| 2 | `Copy` with a directory source fails late (after `MkdirAll`); the "is a directory" error is accurate | S | An early check would add an error shape |
| 3 | `Copy` returns `""` if the fallback link succeeded but removing the temp file failed | S | Unreachable on Linux and macOS now; untestable without fault injection |
| 4 | A custom strategy returning a path in a different directory makes `Copy` stage the temp file in the wrong directory, so it may fail with EXDEV | S | Edge case of user-supplied strategies |
| 5 | When FindAvailableNameTS falls back on an exact-nanosecond collision, the counter lands inside the timestamp | S | Predates this run; cosmetic |
| 6 | Test gaps: link-error routing for errors other than EEXIST, and the remove-failure branch in `renameLink` | M | Needs fault injection, which BRIEF rules out (no new abstractions) |
| 7 | Nit: the `testRenameNoReplace` parameter shadows the package function | XS | Cosmetic |

## Risks to review by hand
1. **seccomp [UNVERIFIED].** If a sandbox or seccomp profile rejects `renameat2` with EPERM instead of ENOSYS, Linux treats it as a real permission error and every Rename and Move fails where it used to fall back. Current Docker and systemd profiles allow `renameat2`. I could not check old container profiles or Android app sandboxes.
2. **Windows and macOS were not run**, only cross-compiled and vetted. The macOS `renamex_np` path and every Windows behaviour depend on CI.
3. **Error identity changed** in the cases listed above. Code that matched `ErrMaxAttemptsReached` for directory destinations will now see `fs.ErrExist`.
4. **Root-only test.** `TestRenameNoReplaceUnprivileged` runs only as root, so it skips in GitHub Actions and the sticky-directory regression isn't guarded in CI.
5. `.ratchet/` is committed on this branch (brief, logs, smoke harness). Drop it before merging if you don't want it in `main`.

## How to reproduce
```bash
.ratchet/check.sh
```
```bash
cd .ratchet/smoke && go run . | diff - expected.txt
```
