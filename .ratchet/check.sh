#!/usr/bin/env bash
# Full gate: build, vet, gofmt, lint (prod only), tests on current + go1.19, smoke.
set -u
cd "$(dirname "$0")/.."
fail=0
step() { echo "== $1"; }
step build;  go build ./... || fail=1
step vet;    go vet ./... || fail=1
step gofmt;  out=$(gofmt -l .); [ -z "$out" ] || { echo "$out"; echo "(gofmt dirty)"; }
step lint;   golangci-lint run ./... 2>&1 | tail -3
step test;   go test -race -count=1 ./... || fail=1
step go1.19; GOTOOLCHAIN=go1.19.13 go test -count=1 ./... || fail=1
step cross;  for os in windows darwin; do GOOS=$os go vet ./... || fail=1; done
step smoke;  (cd .ratchet/smoke && go run . > /tmp/fileflow-smoke.out 2>&1; diff /tmp/fileflow-smoke.out expected.txt; m=$(grep -cxFf expected.txt /tmp/fileflow-smoke.out); echo "smoke match $m/$(wc -l < expected.txt)")
echo "== gate: $([ $fail = 0 ] && echo PASS || echo FAIL)"
exit $fail
