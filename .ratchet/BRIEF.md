# BRIEF (locked 2026-11-04) — law for every pass

## Intent
Go library for programs that file things away (downloads, media import, config
publishing). Move/Copy/Rename never clobber different content: conflicts get a
`-N` suffix, identical destinations dedup (Move/Rename remove src), cross-FS
moves copy+remove. Replace is the explicit atomic overwrite of an exact path.

## Golden path + fixtures
`.ratchet/smoke/` (nested module, `replace` → repo root). Run:
`cd .ratchet/smoke && go run . | diff - expected.txt`.
Covers: plain move, move conflict, move identical (dedup), copy conflict, copy
identical, cross-FS move (/dev/shm), Replace, move onto existing dir, dotfile
conflict, dangling-symlink dst, missing-src side effects. `expected.txt` is the
post-brief expected behavior; baseline match rate is recorded in BASELINE.md.

## Invariants (sacred)
1. Every exported identifier and signature stays as-is. No new exports.
2. `-N` suffix format for ordinary names is unchanged (`a.txt` → `a-1.txt`).
3. Move/Copy/Rename never overwrite a non-identical file.
4. Minimum Go 1.19: build + test must pass with `GOTOOLCHAIN=go1.19.13`.
5. Error types/text may change only where the current one is wrong.

## Decisions
- Q3 (a): Move/Copy/Rename onto an existing directory fail immediately with
  `*ErrFailedMovingFile` matching `fs.ErrExist` (same shape as Replace). No
  into-directory semantics. Replace stays exact-path.
- Dangling symlink at dst = occupied name → conflict suffix.
- Dotfiles: a leading dot is not an extension: `.bashrc` → `.bashrc-1`.
  `x.tar.gz` → `x.tar-1.gz` stays; document it.
- Unix no-replace rename: `renameat2(RENAME_NOREPLACE)` on linux,
  `renamex_np(RENAME_EXCL)` on darwin, via golang.org/x/sys/unix (already in
  go.mod). link+remove remains the fallback.

## Non-goals
New features, new exports, file restructuring, into-directory moves, perf work
off the Equal hot path.

## Appetite: conservative (internals + bug fixes).

## Boundaries
- No new module dependencies. x/sys/unix is allowed (same module).
- CI: may add a `gofmt -l` check only. No other CI/repo-setting changes.
- CRLF → LF for fileflow.go and readme.md: allowed.
- No network/paid APIs in tests or smoke.
- Branch: `claude/cool-faraday-t8kd94` (harness-assigned). Default branch `main`
  is never touched. Push after every sweep. Draft PR at the end, never merged.

## Ranked rocks
1. Destination resolution correctness (dir dst, dangling symlink, dotfiles,
   missing-src dir side effects, Copy path-on-temp-removal-failure).
2. Unix no-replace primitive (protected_hardlinks EPERM → silent overwrite).
3. Equal performance + hygiene (CRLF, gofmt, errcheck, redundant sameFile, CI gofmt).

## Done signal
Dir-dst, dotfile, and EPERM-overwrite bugs fixed, each with a test that failed
first; Equal ≥5× faster on 100 MB; gofmt clean; prod-code errcheck addressed;
full suite green on go1.24 and go1.19; no exported API change.
