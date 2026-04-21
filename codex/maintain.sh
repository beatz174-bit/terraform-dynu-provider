#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./lib.sh
source "$SCRIPT_DIR/lib.sh"

cd_repo_root

log "Refreshing Codex environment and running repository checks"

"$SCRIPT_DIR/setup.sh"

if [[ -x scripts/check.sh ]]; then
  ./scripts/check.sh
else
  log "scripts/check.sh not found; running go test ./..."
  go test ./...
fi

log "Maintenance complete"
