#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=./lib.sh
source "$SCRIPT_DIR/lib.sh"

cd_repo_root

log "Preparing repository-local Codex environment in $(pwd)"
log "Using standalone setup (no external repositories or helper scripts required)"

for cmd in bash git; do
  have "$cmd" || die "$cmd is required"
done

install_terraform() {
  if have terraform; then
    log "terraform detected: $(command -v terraform)"
    return 0
  fi

  have curl || die "curl is required to install terraform"
  have unzip || die "unzip is required to install terraform"

  local tf_version tf_zip tf_url tf_install_dir tf_tmp_dir
  tf_version="${TERRAFORM_VERSION:-1.11.4}"
  tf_zip="terraform_${tf_version}_linux_amd64.zip"
  tf_url="https://releases.hashicorp.com/terraform/${tf_version}/${tf_zip}"
  tf_install_dir="$(pwd)/.codex/bin"
  tf_tmp_dir="$(pwd)/.codex/tmp"

  log "terraform not found; installing v${tf_version} into ${tf_install_dir}"
  curl -fsSL "$tf_url" -o "${tf_tmp_dir}/${tf_zip}"
  unzip -qo "${tf_tmp_dir}/${tf_zip}" -d "$tf_install_dir"
  chmod +x "${tf_install_dir}/terraform"
}

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
    warn "GOPROXY appears malformed: non-empty value contains no entries"
    warn "This can cause errors like: GOPROXY list is not the empty string, but contains no entries"
    warn "Suggested fix: go env -w GOPROXY=https://proxy.golang.org,direct"
    warn "Current GOPROXY from environment: '$value'"
  fi
}

if [[ "${GOPROXY+x}" == "x" ]]; then
  validate_goproxy "$GOPROXY"
fi

ensure_dir .codex
ensure_dir .codex/bin
ensure_dir .codex/cache
ensure_dir .codex/tmp
ensure_dir .codex/logs
export PATH="$(pwd)/.codex/bin:$PATH"

cat > .codex/env <<'ENVEOF'
export CI=1
export TERM=xterm-256color
export GIT_PAGER=cat
export PATH="$(pwd)/.codex/bin:$PATH"
ENVEOF

install_terraform

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
