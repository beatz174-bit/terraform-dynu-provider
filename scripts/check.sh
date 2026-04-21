#!/usr/bin/env bash
set -euo pipefail

echo "[check] Running gofmt"
files="$(git ls-files '*.go')"
if [[ -n "$files" ]]; then
  # shellcheck disable=SC2086
  gofmt -w $files
fi

echo "[check] Running go test"
go test ./...

echo "[check] Completed"
