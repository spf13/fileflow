# Plan (draft, pre-brief)

## Inferred brief
- **Intent:** Go library for moving/copying/renaming files without ever clobbering different content, with an explicit atomic-overwrite escape hatch (`Replace`). Evidence: fileflow.go:17-25 package doc; readme.md "Safety guarantees".
- **Primary user:** Go programs that ingest/organize files (downloads, media import, config publishing — readme Replace example `kennyb.toml`).
- **Golden path:** `Move(src, dst)` → final path; on conflict → `-N` suffix; identical → dedup + remove src; cross-FS → copy+remove. `Copy` same with temp+rename. `Replace` exact-path publish.
- **Invariants (inferred):**
  1. Never overwrite non-identical content in Move/Copy/Rename (readme Safety; fileflow.go:207-210).
  2. Exported API stable — recent commits are already `feat!` breaking (92fc67d, 960b757, cfb7b52), so pre-1.0 tolerance exists, but no tags checked.
  3. Default naming `-N` before extension (fileflow.go:268-281; tests at fileflow_test.go:1023).
  4. go 1.19 minimum (go.mod, CI matrix).
  5. Copies land atomically via temp in dest dir (fileflow.go:428-455).

## Top 3 rocks
1. **Destination resolution correctness** (fileflow.go:179 resolveDst, :262 Exists, :268-338 naming). Evidence: dir dst and dangling-symlink dst spin 100 retries → `ErrMaxAttemptsReached`; dotfile conflict → `-1.bashrc`; Move with missing src creates dirs; Copy reports failure (path "") when the temp link succeeded but temp removal failed (fileflow.go:436-441 vs fileflow_notwindows.go:29-31).
   Lenses: completeness → correctness → tests → simplify.
2. **Unix no-replace primitive** (fileflow_notwindows.go:26-48). link+remove falls back to plain `os.Rename` on EPERM — which Linux returns under `fs.protected_hardlinks=1` (=1 here) when linking a file you don't own (e.g. shared /tmp), silently voiding the no-overwrite guarantee. Use `renameat2(RENAME_NOREPLACE)` (linux) / `renamex_np(RENAME_EXCL)` (darwin) from already-required x/sys, keep link+remove as fallback. Lenses: correctness → security → architecture → tests.
3. **Equal perf + hygiene** (fileflow.go:341-392). Byte-at-a-time compare ≈ 8.6 ms/MB; chunked `bytes.Equal` should be ~10×+. Plus: fileflow.go is CRLF (gofmt -l fails), replace.go gofmt, 4 errcheck in prod code, redundant `sameFile` in Move (fileflow.go:160, repeated in resolveDst), CI lacks gofmt. Lenses: performance → simplify → docs → tests.

## Findings backlog (ranked)
| # | Lens | Finding | Evidence |
|---|---|---|---|
| 1 | correctness | Dir/dangling-symlink dst → 100 futile retries, misleading error | smoke; fileflow.go:183,240 |
| 2 | security | EPERM fallback to overwriting os.Rename under protected_hardlinks | fileflow_notwindows.go:542 |
| 3 | correctness | Dotfile conflict name `-1.bashrc` | fileflow.go:270-274 (filepath.Ext on ".bashrc") |
| 4 | correctness | Copy returns ("", err) though content committed when temp removal fails | fileflow.go:436-441 |
| 5 | performance | Equal byte-at-a-time | fileflow.go:371-392; smoke 865ms/100MB |
| 6 | completeness | Move/Rename/Copy with missing src create dest dirs before failing | fileflow.go:225,414; smoke |
| 7 | docs/tooling | CRLF fileflow.go + readme.md; gofmt not clean; CI no gofmt | git ls-files --eol |
| 8 | simplify | Move's sameFile check duplicates resolveDst | fileflow.go:160 |
| 9 | tests | Move cross-FS fallback (fileMove) untested on Linux; Error() strings 0% cov | coverage |
| 10 | correctness | Custom strategy returning a path in another dir: Copy stages temp in dst dir, rename may EXDEV | fileflow.go:428 |
| 11 | docs | `x.tar.gz` → `x.tar-1.gz` (document or fix) | smoke |
| 12 | lint | errcheck on deferred cleanup (intentional; annotate or `_ =`) | golangci-lint |
