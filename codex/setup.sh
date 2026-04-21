#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./lib.sh
source "$SCRIPT_DIR/lib.sh"

cd_repo_root

log "Preparing repository-local Codex environment in $(pwd)"

for cmd in bash git; do
  have "$cmd" || die "$cmd is required"
done

ensure_dir .codex
ensure_dir .codex/bin
ensure_dir .codex/cache
ensure_dir .codex/tmp
ensure_dir .codex/logs

cat > .codex/env <<'ENVEOF'
export CI=1
export TERM=xterm-256color
export GIT_PAGER=cat
ENVEOF

if have python3; then
  log "python3 detected"
  if [[ ! -d .codex/venv ]]; then
    python3 -m venv .codex/venv
  fi
  # shellcheck disable=SC1091
  source .codex/venv/bin/activate
  python -m pip install --upgrade pip setuptools wheel >/dev/null
else
  warn "python3 not found; skipping optional venv setup"
fi

find codex -maxdepth 1 -type f -name '*.sh' -exec chmod +x {} +

touch .git/info/exclude
append_if_missing ".codex/" .git/info/exclude
append_if_missing ".pytest_cache/" .git/info/exclude
append_if_missing "__pycache__/" .git/info/exclude

log "Setup complete"
