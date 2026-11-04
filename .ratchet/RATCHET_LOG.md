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

### S1 · completeness — KEPT — 55→66 — fdd91fc
- Critique: missing src created dst parent dirs and, onto an existing dir, was misreported as ErrExist (fileflow.go:225,414).
- Change: Lstat (Rename) / Stat (Copy) src before any destination work.
- Verified: TestMissingSourceHasNoSideEffects failed first (3 ops); check.sh PASS; smoke 12→13/14.
- Judge: KEEP. Minor residue (backlog): Rename(m,m)/Copy(m,m) on a missing path now ErrNotExist while Move(m,m) is ErrSameFile; cross-FS Move of a dangling-symlink src still mkdirs; Copy of a directory src fails late.

### S1 · correctness(naming) — KEPT — 55→85 — 38a057d
- Critique: filepath.Ext(".bashrc") == ".bashrc" → every strategy produced "-1.bashrc" (fileflow.go:270 et al.).
- Change: splitExt helper (leading dot belongs to the name), replaces 4 copies of the split.
- Verified: TestFindAvailableNameDotfiles failed first (8 asserts); check.sh PASS; smoke 13→14/14.
- Judge: KEEP; 30-input old/new comparison found no change for ordinary names. Docs gap → docs lens.

### S1 · performance — KEPT — 20→90 — 7f6da45
- Critique: readersEqual compared byte-by-byte via bufio.ReadByte (~105 MB/s; 100 MB in ~865 ms).
- Change: io.ReadFull 64 KiB chunks + bytes.Equal; BenchmarkEqual; chunk-boundary tests.
- Verified: BenchmarkEqual 10 MiB 96 ms → 2.4 ms (~40×, implementer) / ~30× (judge); 100 MB 865 → 38 ms; check.sh PASS.
- Judge: KEEP; 3000 randomized cases vs reference, mutation (ReadFull→Read) caught by tests. Minor: a reader returning (0,nil) forever would now spin (old bufio gave ErrNoProgress) — unreachable with *os.File.

### S1 · simplify — KEPT — 72→76 — 84af3c7
- Critique: Move's sameFile duplicated resolveDst's (fileflow.go:160).
- Change: deleted it. Move(p,p) with missing p now ErrNotExist (consistent with Rename/Copy).
- Verified: check.sh PASS, smoke 14/14.
- Judge: KEEP; alias matrix (spellings, hard/sym links, cross-device) identical before/after.

### S1 · architecture(lint) — KEPT — 55→80 — 73c1085
- Critique: golangci-lint errcheck 11 shown (12 real; default cap hid replace.go:58).
- Change: explicit `_ =` on best-effort cleanup; benchmark uses b.TempDir and checks Remove.
- Verified: golangci-lint 0 issues; check.sh PASS.
- Judge: KEEP; written temp file's Sync/Close are checked on the success path, every ignored error is ignorable.

### S1 · ci(gofmt) — KEPT — 50→85 — 53231c8
- Change: one `gofmt -l` step on ubuntu/stable only (BRIEF allows exactly this).
- Judge: KEEP; condition matches matrix strings; bash syntax valid.
