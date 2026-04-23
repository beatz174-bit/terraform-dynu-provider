#!/usr/bin/env bash
set -euo pipefail

echo "[check] verifying gofmt"
files="$(git ls-files '*.go')"
if [[ -n "${files}" ]]; then
  unformatted="$(gofmt -l ${files})"
  if [[ -n "${unformatted}" ]]; then
    echo "[check][error] gofmt reported unformatted files:" >&2
    echo "${unformatted}" >&2
    exit 1
  fi
fi

echo "[check] running go vet"
go vet ./...

echo "[check] running go test"
go test ./...

if command -v terraform >/dev/null 2>&1; then
  echo "[check] running terraform fmt check for examples/"
  terraform fmt -check -recursive examples
else
  echo "[check][warn] terraform not found; skipping examples/ checks" >&2
fi

echo "[check] complete"
