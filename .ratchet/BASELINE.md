# Baseline — 2026-11-04 @ 4be9157 (branch claude/cool-faraday-t8kd94)

Toolchain: go1.24.7 linux/amd64, golangci-lint 2.5.0. Running as root.

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | PASS (~8.3s cold, incl. module download) |
| Vet | `go vet ./...` | PASS, 0 findings |
| Tests | `go test -race -count=1 ./...` | PASS |
| Test count | `go test -v` top-level + subtests | 93 RUN/subtest lines; 1 SKIP (`TestReplaceCrossFilesystemSourceRemovalFailure`: root) |
| Coverage | `-coverprofile` | 83.4% statements |
| gofmt | `gofmt -l .` | **FAIL**: `fileflow.go` (CRLF line endings, whole file), `replace.go` (missing blank line at :67) |
| Lint | `golangci-lint run` (default v2 linters) | 11 errcheck (4 in fileflow.go deferred Close/Remove, rest in tests) |
| Bench | `go test -bench=. -benchmem` | BenchmarkCopy 10MB: ~22.9 ms/op, 2.4 KB/op, 33 allocs/op |
| Size | `wc -l *.go` | 2688 lines total (fileflow.go 503, replace*.go ~120 non-test) |
| CI | `.github/workflows/test.yml` | vet + `test -race` on ubuntu/macos/windows × go1.19/stable. No gofmt, no lint. |

## Smoke (library; no CLI) — scratch program against local module
- `Move(a.txt, existingDir)` → `"" maximum increment attempts reached` (100 futile retries)
- `Copy(b, s/.bashrc)` with conflict → `s/-1.bashrc`
- `Move(c, danglingSymlink)` → `"" maximum increment attempts reached`
- `Move(missing, new/deep/x)` → error, but `new/deep/` was created as a side effect
- `Equal` 100 MB identical files → ~865 ms (~8.6 ms/MB, byte-at-a-time)
- `Copy(t2, t/x.tar.gz)` with conflict → `x.tar-1.gz`

## Could not run
- Windows / macOS paths (no host). Source-removal-failure test (root).

## Post-brief additions
- go1.19.13 (`GOTOOLCHAIN=go1.19.13 go test ./...`): PASS. `GOOS=windows|darwin go vet`: PASS.
- Golden smoke (`.ratchet/smoke`, gate `.ratchet/check.sh`): **8/14** lines match expected.txt.
  Mismatches: 3× onto-dir (ErrMaxAttemptsReached), dotfile (`-1.bashrc`), dangling dst, missing-src creates dirs.
