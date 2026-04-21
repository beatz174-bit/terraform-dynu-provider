#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./lib.sh
source "$SCRIPT_DIR/lib.sh"

cd_repo_root

echo "=== codex doctor ==="
echo "repo: $(pwd)"
echo "time: $(date -u +%FT%TZ)"
echo

echo "--- tools ---"
for cmd in bash git go terraform; do
  if have "$cmd"; then
    echo "$cmd: $(command -v "$cmd")"
  else
    echo "$cmd: MISSING"
  fi
done
echo

echo "--- repo ---"
for path in README.md go.mod main.go internal/provider/provider.go; do
  if [[ -f "$path" ]]; then
    echo "$path: present"
  else
    echo "$path: missing"
  fi
done
echo

echo "--- git status ---"
git status --short || true
