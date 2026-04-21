#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./lib.sh
source "$SCRIPT_DIR/lib.sh"

cd_repo_root

log "Starting Codex environment setup in $(pwd)"

if ! have git; then
  die "git is required"
fi

if ! have bash; then
  die "bash is required"
fi

# Keep all Codex-local state inside the repo or user home.
ensure_dir .codex
ensure_dir .codex/bin
ensure_dir .codex/cache
ensure_dir .codex/tmp
ensure_dir .codex/logs

# Useful environment defaults for repeatable non-interactive runs.
cat > .codex/env <<'EOF'
export CI=1
export TERM=xterm-256color
export PYTHONDONTWRITEBYTECODE=1
export PIP_DISABLE_PIP_VERSION_CHECK=1
export PIP_NO_INPUT=1
export GIT_PAGER=cat
EOF

# Optional Python venv for helper tooling.
if have python3; then
  log "python3 detected"
  if [[ ! -d .codex/venv ]]; then
    python3 -m venv .codex/venv
  fi
  # shellcheck disable=SC1091
  source .codex/venv/bin/activate
  python -m pip install --upgrade pip setuptools wheel >/dev/null

  # Install only lightweight tooling that is broadly useful.
  python -m pip install \
    pyyaml \
    requests \
    jinja2 \
    >/dev/null
else
  warn "python3 not found; skipping virtualenv setup"
fi

# Basic repo checks based on your environment style.
if [[ -f services-up.sh ]]; then
  log "Found services-up.sh"
  chmod +x services-up.sh || true
else
  warn "services-up.sh not found at repo root"
fi

# Make codex scripts executable.
find codex -maxdepth 1 -type f -name '*.sh' -exec chmod +x {} +

# Create a simple local ignore file for Codex-generated junk if needed.
touch .git/info/exclude
append_if_missing ".codex/" .git/info/exclude
append_if_missing ".pytest_cache/" .git/info/exclude
append_if_missing "__pycache__/" .git/info/exclude

# Helpful summary file for Codex tasks.
cat > .codex/README.md <<'EOF'
This directory contains local, disposable Codex environment state.

Contents:
- env: shell exports for non-interactive work
- venv: optional Python virtual environment
- cache/: tool cache
- tmp/: temporary files
- logs/: script logs
EOF

log "Setup complete"
