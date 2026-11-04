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

### S1 · docs — KEPT — 70→84 — 8bde877
- Critique: README/doc comments silent on dir dst, dangling symlinks, dotfiles; "atomic link+remove on Unix" stale.
- Change: README safety/naming/errors sections, Rename + FindAvailableNameInc doc comments.
- Verified: README full example and Flow/custom/Replace snippets compile (go vet); check.sh PASS.
- Judge: KEEP; all claims verified incl. strace of renameat2. Follow-ups (→S2 docs): readme.md:78 stale "behavior is unchanged"; custom-strategy example uses Exists/filepath.Ext (misses occupied non-files, dotfiles); Replace missing from ErrFailedMovingFile list; Copy/Move doc comments lack dir note.

### S1 · correctness/security(rename) — KEPT — 60→85 — 03d1959
- Critique: link+remove isn't atomic: unremovable src (sticky dir) left dst as a hard-link alias of another user's file; EPERM/ENOTSUP fallback to os.Rename can overwrite in a race window (fileflow_notwindows.go:26-48).
- Change: renameat2(RENAME_NOREPLACE) on linux/android, renamex_np(RENAME_EXCL) on darwin/ios; link+remove fallback elsewhere/unsupported.
- Verified: TestRenameNoReplaceUnprivileged/source_not_removable failed first (alias left behind); check.sh PASS; cross-builds.
- Judge: KEEP. Correction to Phase C claim: the EPERM overwrite is race-only (kernel returns EEXIST before may_linkat EPERM) — untested. Test fixes (skip-on-exec-failure, inverted comment) → 8579d37. Risk [UNVERIFIED]: a seccomp profile returning EPERM for renameat2 would now fail renames instead of falling back.

### S1 · tests — KEPT — 40→72 — 8579d37
- Critique: after 03d1959 the link+remove fallback was unreachable in tests on Linux; coverage 83.4→82.9%. Helper test failed instead of skipping when it could not exec.
- Change: extracted renameLink; shared testRenameNoReplace runs against it on !windows; skip only on exec start failure.
- Verified: coverage 84.9%; skip confirmed under 0700 TMPDIR; check.sh PASS.
- Judge: KEEP; mutations "always os.Rename" and "skip remove" caught (old tests caught neither). Gaps (backlog): non-EEXIST link-error routing and remove-failure branch untested without fault injection.

## Sweep 1 scorecard
| Lens | Result | Score |
|---|---|---|
| correctness (occupied names) | KEPT 8b045a2 | 55→72 |
| completeness (missing src) | KEPT fdd91fc | 55→66 |
| correctness (dotfiles) | KEPT 38a057d | 55→85 |
| correctness/security (atomic rename) | KEPT 03d1959 | 60→85 |
| performance (Equal) | KEPT 7f6da45 | 20→90 |
| simplify (Move) | KEPT 84af3c7 | 72→76 |
| docs | KEPT 8bde877 | 70→84 |
| architecture (lint) | KEPT 73c1085 | 55→80 |
| ci (gofmt) | KEPT 53231c8 | 50→85 |
| tests | KEPT 8579d37 | 40→72 |
| reality (lens 15) | PASS — check.sh at 8579d37 | — |

| Metric | Baseline | After S1 |
|---|---|---|
| Tests (RUN+subtests) | 93 | 139 |
| Coverage | 83.4% | 84.9% |
| golangci-lint | 11 (12 real) | 0 |
| gofmt -l | 2 files | 0 |
| Smoke match | 8/14 | 14/14 |
| Equal 100 MB | ~865 ms | ~38 ms |
| BenchmarkCopy 10 MB | ~22.9 ms | ~23.4 ms (noise) |
| Cold build | ~8.3 s | ~7.8 s |
| Prod Go lines | 714 | 825 |

### S2 · docs — KEPT — 72→82 — 20e721f
- Critique (from S1 docs judge): readme.md:78 stale "behavior is unchanged"; custom-strategy example used Exists (misses dirs/dangling links); Replace errors unlisted; Copy/Move doc comments lacked dir note.
- Change: fixed each; README snippets re-compiled with go vet.
- Judge: KEEP; example verified to skip dir/dangling candidates. Remaining pre-existing falsehood: Exists "regular file" → bookend.

### S2 · completeness — CONVERGED — 66→66
- Remaining items are low-impact or untestable without fault injection: Copy of a directory src fails late (after mkdir) with an accurate "is a directory" error; Copy returns "" if the fallback link succeeds but temp removal fails (unreachable on linux/darwin now); custom strategies returning another directory make Copy stage in the wrong dir. Not worth churn.

### S2 · correctness — CONVERGED — 85→85
- No reproducible defect left. Seccomp-EPERM-on-renameat2 risk recorded for human review.

### S2 · tests — CONVERGED — 72→72
- Remaining gaps need fault injection (BRIEF: no new abstractions). Nit (not changed, cosmetic): fileflow_test.go:79 param shadows renameNoReplace.

### Bookend · architecture / simplify — CONVERGED
- Re-read fileflow.go end to end. Candidates (inline fileMove; drop Equal's always-true final condition) are cosmetic; no deletion with real value left.

### Bookend · ux / design (library: error text, godoc) — CONVERGED
- Errors name the operation, both paths, and wrap the cause (e.g. "failed moving file a to dir: destination is a directory: file already exists").

### Bookend · docs — aed4256 (judged in final gate)
- Exists "regular file" was false (true for devices/FIFOs); ErrMaxAttemptsReached described too narrowly.

## Sweep 2 scorecard
| Lens | Result | Score |
|---|---|---|
| docs | KEPT 20e721f | 72→82 |
| completeness | CONVERGED | 66 |
| correctness | CONVERGED | 85 |
| tests | CONVERGED | 72 |
| reality | PASS (check.sh at 20e721f) | — |
Convergence: sweep 2 → 3 of 4 lenses CONVERGED (75%) + bookend all CONVERGED; no high-impact backlog item remains; done signal met (see REPORT.md). Stopping.
- Correction: 75% (sweep 2) / 78% (sweep 2 + bookend) is below the 80% convergence threshold. The stop condition that applies is "BRIEF.md's done signal is met and verified" (verified independently by the final reality gate).
