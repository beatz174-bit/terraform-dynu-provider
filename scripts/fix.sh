#!/usr/bin/env bash
set -euo pipefail

if [[ -x codex/setup.sh ]]; then
  ./codex/setup.sh >/dev/null
fi

if [[ -d .codex/bin ]]; then
  export PATH="$(pwd)/.codex/bin:$PATH"
fi

echo "[fix] formatting Go files"
files="$(git ls-files '*.go')"
if [[ -n "${files}" ]]; then
  gofmt -w ${files}
fi

echo "[fix] running go mod tidy"
go mod tidy

if command -v terraform >/dev/null 2>&1; then
  echo "[fix] running terraform fmt for examples/"
  terraform fmt -recursive examples
else
  echo "[fix][warn] terraform not found; skipping examples/ formatting" >&2
fi

echo "[fix] complete"
