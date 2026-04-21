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

validate_goproxy() {
  local value="$1"
  local part trimmed
  local has_valid=0

  IFS=',' read -r -a parts <<<"$value"
  for part in "${parts[@]}"; do
    trimmed="${part#"${part%%[![:space:]]*}"}"
    trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
    if [[ -n "$trimmed" ]]; then
      has_valid=1
      break
    fi
  done

  if [[ "$has_valid" -eq 0 ]]; then
    echo "[setup-dev][warn] GOPROXY appears malformed: non-empty value contains no entries." >&2
    echo "[setup-dev][warn] This can cause errors like: GOPROXY list is not the empty string, but contains no entries." >&2
    echo "[setup-dev][warn] Suggested fix: go env -w GOPROXY=https://proxy.golang.org,direct" >&2
    echo "[setup-dev][warn] Current GOPROXY from environment: '$value'" >&2
  fi
}

if [[ "${GOPROXY+x}" == "x" ]]; then
  validate_goproxy "$GOPROXY"
fi

if command -v terraform >/dev/null 2>&1; then
  echo "[setup-dev] terraform: $(command -v terraform)"
else
  echo "[setup-dev][warn] terraform not found (only required for terraform validate and acceptance workflows)"
fi

echo "[setup-dev] Toolchain validation complete"
