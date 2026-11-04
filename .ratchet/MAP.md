# Map
Archetype: **engine** (Go library, no CLI/UI). Module `github.com/spf13/fileflow`, go 1.19, dep `golang.org/x/sys` (windows only today).

## Public API (all in package fileflow)
- fileflow.go: `Move`, `Rename`, `Copy` (pkg funcs + `*Flow` methods), `Exists`, `Equal`,
  `FindAvailableName{Inc,Cont,Auto,TS}`, `Flow{FindAvailableName, DirMode, NoCreateDirs}`,
  `DefaultDirMode`, `ErrSameFile`, `ErrMaxAttemptsReached`, `ErrFailedRemovingOriginal`, `ErrFailedMovingFile`
- replace.go: `Replace`, `(*Flow).Replace`

## Core internals
- fileflow.go:179 `resolveDst` — conflict decision (sameFile → Exists → Equal → naming)
- fileflow.go:211 `Rename` retry loop; :413 `Copy` retry loop; :459 `copyToTemp`
- fileflow.go:341 `Equal` / :371 `readersEqual` (bufio ReadByte loop)
- fileflow_notwindows.go:26 `renameNoReplace` = link+remove, EPERM/ENOTSUP fallback to os.Rename
- fileflow_windows.go:20 `renameNoReplace` = MoveFileEx(0)
- replace_{notwindows,windows}.go `renameReplace`

## Tests
- fileflow_test.go (1500 lines, 40 tests) — ops, naming strategies, races, Flow config
- replace_test.go (8 tests), replace_crossdevice_test.go (linux, uses /dev/shm as second FS)
- fileflow_bench_test.go — BenchmarkCopy 10 MB
- No fixtures dir; all fixtures created in t.TempDir()

## Churn hotspots (commits touching)
fileflow.go 28 · fileflow_test.go 21 · readme.md 15 · platform files 3 each · replace* 1

## Surfaces users see
README (CRLF), godoc comments, error strings, returned final paths.
