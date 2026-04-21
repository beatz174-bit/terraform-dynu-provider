#!/usr/bin/env bash
set -euo pipefail

echo "[setup-dev] Validating local toolchain"

for cmd in bash git go; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "[setup-dev][error] Missing required command: $cmd" >&2
    exit 1
  fi
  echo "[setup-dev] $cmd: $(command -v "$cmd")"
done

if command -v terraform >/dev/null 2>&1; then
  echo "[setup-dev] terraform: $(command -v terraform)"
else
  echo "[setup-dev][warn] terraform not found (only required for terraform validate and acceptance workflows)"
fi

echo "[setup-dev] Toolchain validation complete"
