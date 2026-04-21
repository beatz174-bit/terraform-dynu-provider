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
for cmd in bash git python3 docker; do
  if have "$cmd"; then
    echo "$cmd: $(command -v "$cmd")"
  else
    echo "$cmd: MISSING"
  fi
done
echo

echo "--- repo ---"
[[ -f services-up.sh ]] && echo "services-up.sh: present" || echo "services-up.sh: missing"
[[ -d core ]] && echo "core/: present" || true
[[ -d apps ]] && echo "apps/: present" || true
[[ -d monitoring ]] && echo "monitoring/: present" || true
echo

echo "--- git status ---"
git status --short || true
echo

if have docker && docker compose version >/dev/null 2>&1; then
  echo "--- docker compose ---"
  docker compose version || true
fi
