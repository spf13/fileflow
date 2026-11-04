# Ratchet log (append-only)

Branch: claude/cool-faraday-t8kd94 (harness-assigned; default branch `main` has no origin/HEAD symref, fallback used).
Decision: implementer = main session; judge = fresh subagent per pass that sees only BRIEF, BASELINE, diff, touched files.

### S1 · prep(format) — KEPT — n/a — 47912f9
- Critique: fileflow.go/readme.md CRLF; `gofmt -l` flagged fileflow.go (whole file) and replace.go:67.
- Change: CRLF→LF (BRIEF allows), gofmt replace.go. Whitespace only (`git diff -w` = 1 blank line).
- Verified: `gofmt -l .` empty; check.sh PASS.
- Judge: not judged separately (objective, whitespace-only); decision logged.

### S1 · correctness — KEPT — 55→72 — 8b045a2
- Critique: dir / symlink-to-dir / dangling-symlink dst invisible to Exists → 100 futile retries, ErrMaxAttemptsReached (fileflow.go:183,240); strategies could pick a directory-held name.
- Change: `taken()` (Lstat) for occupancy in resolveDst + Inc/Cont/TS; directory dst → `*ErrFailedMovingFile` wrapping fs.ErrExist; Equal only for regular files.
- Verified: 3 new tests failed first (Onto dir ×3 ops, dangling symlink, strategies skip dir); check.sh PASS (go1.24 race, go1.19, win/darwin vet); smoke 8→12/14.
- Judge: KEEP — real fix, tests fail on old code. Minor: missing src onto existing dir now says ErrExist not ErrNotExist (fixed next pass); docs not yet updated (docs lens).
