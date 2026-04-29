#!/usr/bin/env bash
set -euo pipefail

MIN_GO_VERSION="1.23.0"
MIN_TERRAFORM_VERSION="1.5.0"

FIX_MODE=0
STRICT_MODE=1

log() {
  echo "[setup-dev] $*"
}

warn() {
  echo "[setup-dev][warn] $*"
}

error() {
  echo "[setup-dev][error] $*" >&2
}

usage() {
  cat <<USAGE
Usage: ./scripts/setup-dev.sh [--fix] [--no-strict] [--help]

Modes:
  --fix     Validate and attempt safe remediation via already-installed version managers.
  --no-strict  Allow missing Terraform (not recommended; legacy compatibility mode).
  --help    Show this help text.
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --fix)
      FIX_MODE=1
      ;;
    --no-strict)
      STRICT_MODE=0
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      error "Unknown argument: $1"
      usage
      exit 2
      ;;
  esac
  shift
done

is_vscode_terminal() {
  [[ "${TERM_PROGRAM:-}" == "vscode" || -n "${VSCODE_PID:-}" || -n "${VSCODE_CWD:-}" ]]
}

normalize_version() {
  local raw="$1"
  raw="${raw#v}"
  raw="${raw%%-*}"
  raw="${raw%%+*}"
  local major minor patch
  IFS='.' read -r major minor patch <<<"$raw"
  major="${major:-0}"
  minor="${minor:-0}"
  patch="${patch:-0}"
  printf '%s.%s.%s\n' "$major" "$minor" "$patch"
}

version_ge() {
  local a b
  a="$(normalize_version "$1")"
  b="$(normalize_version "$2")"
  local a1 a2 a3 b1 b2 b3
  IFS='.' read -r a1 a2 a3 <<<"$a"
  IFS='.' read -r b1 b2 b3 <<<"$b"

  if (( a1 > b1 )); then return 0; fi
  if (( a1 < b1 )); then return 1; fi
  if (( a2 > b2 )); then return 0; fi
  if (( a2 < b2 )); then return 1; fi
  if (( a3 >= b3 )); then return 0; fi
  return 1
}

command_path() {
  command -v "$1" 2>/dev/null || true
}

check_required_command() {
  local cmd="$1"
  local path
  path="$(command_path "$cmd")"
  if [[ -z "$path" ]]; then
    error "Missing required command: $cmd"
    return 1
  fi
  log "$cmd path: $path"
  return 0
}

parse_go_version() {
  local output="$1"
  sed -nE 's/^go version go([0-9]+(\.[0-9]+){1,2}).*/\1/p' <<<"$output"
}

parse_tf_version_text() {
  local output="$1"
  sed -nE 's/^Terraform v([0-9]+(\.[0-9]+){1,2}).*/\1/p' <<<"$output" | head -n1
}

parse_tf_version_json() {
  local output="$1"
  sed -nE 's/.*"terraform_version"[[:space:]]*:[[:space:]]*"([0-9]+(\.[0-9]+){1,2})".*/\1/p' <<<"$output" | head -n1
}

has_go_stdlib_package() {
  local goroot="$1"
  local pkg="$2"
  [[ -d "$goroot/src/$pkg" ]]
}

validate_goproxy_entry() {
  local entry="$1"
  if [[ -z "$entry" ]]; then
    return 1
  fi
  case "$entry" in
    off|direct)
      return 0
      ;;
    http://*|https://*|file://*)
      return 0
      ;;
  esac
  return 1
}

check_goproxy() {
  local bad=0
  local env_goproxy goenv_goproxy
  env_goproxy="${GOPROXY:-}"
  goenv_goproxy="$(go env GOPROXY 2>/dev/null || true)"

  if [[ -n "$env_goproxy" ]]; then
    log "GOPROXY env: $env_goproxy"
  fi
  if [[ -n "$goenv_goproxy" ]]; then
    log "go env GOPROXY: $goenv_goproxy"
  fi

  local effective="$goenv_goproxy"
  if [[ -z "$effective" ]]; then
    return 0
  fi

  local valid=0
  IFS=',' read -r -a entries <<<"$effective"
  for entry in "${entries[@]}"; do
    entry="$(echo "$entry" | xargs)"
    if validate_goproxy_entry "$entry"; then
      valid=1
      break
    fi
  done

  if (( valid == 0 )); then
    warn "GOPROXY appears malformed: '$effective'"
    warn "Recommended fix: go env -w GOPROXY=https://proxy.golang.org,direct"
    bad=1
  fi

  return "$bad"
}

get_go_state() {
  GO_PATH="$(command_path go)"
  GO_RAW_VERSION=""
  GO_VERSION=""
  GO_GOROOT=""

  if [[ -n "$GO_PATH" ]]; then
    GO_RAW_VERSION="$(go version 2>/dev/null || true)"
    GO_VERSION="$(parse_go_version "$GO_RAW_VERSION")"
    GO_GOROOT="$(go env GOROOT 2>/dev/null || true)"
  fi
}

validate_go_environment() {
  local bad=0
  get_go_state

  if [[ -z "$GO_PATH" ]]; then
    error "go not found in PATH"
    bad=1
    return "$bad"
  fi

  log "go path: $GO_PATH"
  if [[ -n "$GO_RAW_VERSION" ]]; then
    log "go version: $GO_RAW_VERSION"
  fi
  if [[ -n "$GO_GOROOT" ]]; then
    log "go env GOROOT: $GO_GOROOT"
  fi
  if [[ -n "${GOROOT:-}" ]]; then
    log "GOROOT env: ${GOROOT}"
  fi

  if [[ -z "$GO_VERSION" ]]; then
    error "Unable to parse Go version from: '${GO_RAW_VERSION:-<empty>}'"
    bad=1
  elif version_ge "$GO_VERSION" "$MIN_GO_VERSION"; then
    log "Go version check: PASS (detected $GO_VERSION, required >= $MIN_GO_VERSION)"
  else
    error "Go version check: FAIL (detected $GO_VERSION, required >= $MIN_GO_VERSION)"
    bad=1
  fi

  if [[ -n "${GOROOT:-}" && -n "$GO_GOROOT" && "$GOROOT" != "$GO_GOROOT" ]]; then
    error "GOROOT mismatch: GOROOT env differs from 'go env GOROOT'."
    warn "Usually GOROOT should not be set manually. Try: unset GOROOT"
    bad=1
  fi

  if [[ -n "$GO_GOROOT" ]]; then
    local missing_std=0
    local pkg
    for pkg in "slices" "maps" "math/rand/v2"; do
      if ! has_go_stdlib_package "$GO_GOROOT" "$pkg"; then
        warn "Stdlib package missing from GOROOT ($GO_GOROOT): $pkg"
        missing_std=1
      fi
    done

    if (( missing_std == 1 )); then
      error "Detected a possibly stale or mismatched GOROOT. This can cause build failures on modern Go versions."
      warn "Likely cause: manual GOROOT override or a broken Go installation path precedence."
      warn "Next steps: unset GOROOT, ensure correct 'go' is first in PATH, then re-open shell/terminal."
      bad=1
    fi
  fi

  check_goproxy || true

  return "$bad"
}

get_terraform_state() {
  TERRAFORM_PATH="$(command_path terraform)"
  TERRAFORM_VERSION=""
  TERRAFORM_RAW_VERSION=""

  if [[ -n "$TERRAFORM_PATH" ]]; then
    local json_output
    json_output="$(terraform version -json 2>/dev/null || true)"
    if [[ -n "$json_output" ]]; then
      TERRAFORM_VERSION="$(parse_tf_version_json "$json_output")"
    fi

    TERRAFORM_RAW_VERSION="$(terraform version 2>/dev/null || true)"
    if [[ -z "$TERRAFORM_VERSION" ]]; then
      TERRAFORM_VERSION="$(parse_tf_version_text "$TERRAFORM_RAW_VERSION")"
    fi
  fi
}

validate_terraform() {
  local bad=0
  get_terraform_state

  if [[ -z "$TERRAFORM_PATH" ]]; then
    if (( STRICT_MODE == 1 )); then
      error "terraform not found in PATH (required in --strict mode)"
    else
      warn "terraform not found in PATH (allowed only because --no-strict was set)"
    fi
    bad=1
    return "$bad"
  fi

  log "terraform path: $TERRAFORM_PATH"
  if [[ -n "$TERRAFORM_VERSION" ]]; then
    log "terraform version: $TERRAFORM_VERSION"
  else
    warn "Unable to parse Terraform version from output: ${TERRAFORM_RAW_VERSION//$'\n'/ | }"
  fi

  if [[ -z "$TERRAFORM_VERSION" ]]; then
    bad=1
  elif version_ge "$TERRAFORM_VERSION" "$MIN_TERRAFORM_VERSION"; then
    log "Terraform version check: PASS (detected $TERRAFORM_VERSION, required >= $MIN_TERRAFORM_VERSION)"
  else
    error "Terraform version check: FAIL (detected $TERRAFORM_VERSION, required >= $MIN_TERRAFORM_VERSION)"
    bad=1
  fi

  return "$bad"
}

has_manager() {
  command -v "$1" >/dev/null 2>&1
}

attempt_fix_go() {
  local target="${MIN_GO_VERSION%.*}"
  log "Attempting Go remediation (minimum: $MIN_GO_VERSION)"

  if has_manager mise; then
    log "Using mise for Go remediation"
    if mise install "go@$target" >/dev/null 2>&1 && mise use -q "go@$target" >/dev/null 2>&1; then
      hash -r
      log "mise applied go@$target for this repository"
      return 0
    fi
    warn "mise was detected but automatic Go remediation failed"
  fi

  if has_manager asdf; then
    log "Using asdf for Go remediation"
    if asdf plugin list 2>/dev/null | grep -qx 'golang'; then
      if asdf install golang "$target" >/dev/null 2>&1 && asdf set golang "$target" >/dev/null 2>&1; then
        hash -r
        log "asdf applied golang $target for this repository"
        return 0
      fi
      warn "asdf golang plugin exists but install/set failed"
    else
      warn "asdf detected, but golang plugin is not installed"
      warn "Install plugin manually: asdf plugin add golang https://github.com/asdf-community/asdf-golang.git"
    fi
  fi

  warn "Could not auto-remediate Go with installed managers."
  warn "Manual options:"
  warn "  - mise install go@$target && mise use go@$target"
  warn "  - asdf install golang $target && asdf set golang $target"
  return 1
}

attempt_fix_terraform() {
  local target="${MIN_TERRAFORM_VERSION%.*}"
  log "Attempting Terraform remediation (minimum: $MIN_TERRAFORM_VERSION)"

  if has_manager mise; then
    log "Using mise for Terraform remediation"
    if mise install "terraform@$target" >/dev/null 2>&1 && mise use -q "terraform@$target" >/dev/null 2>&1; then
      hash -r
      log "mise applied terraform@$target for this repository"
      return 0
    fi
    warn "mise was detected but automatic Terraform remediation failed"
  fi

  if has_manager asdf; then
    log "Using asdf for Terraform remediation"
    if asdf plugin list 2>/dev/null | grep -qx 'terraform'; then
      if asdf install terraform "$target" >/dev/null 2>&1 && asdf set terraform "$target" >/dev/null 2>&1; then
        hash -r
        log "asdf applied terraform $target for this repository"
        return 0
      fi
      warn "asdf terraform plugin exists but install/set failed"
    else
      warn "asdf detected, but terraform plugin is not installed"
      warn "Install plugin manually: asdf plugin add terraform https://github.com/asdf-community/asdf-hashicorp.git"
    fi
  fi

  if has_manager tfenv; then
    log "Using tfenv for Terraform remediation"
    if tfenv install "$target" >/dev/null 2>&1 && tfenv use "$target" >/dev/null 2>&1; then
      hash -r
      log "tfenv selected Terraform $target"
      return 0
    fi
    warn "tfenv was detected but install/use failed"
  fi

  if has_manager tenv; then
    log "Using tenv for Terraform remediation"
    if tenv tf install "$target" >/dev/null 2>&1 && tenv tf use "$target" >/dev/null 2>&1; then
      hash -r
      log "tenv selected Terraform $target"
      return 0
    fi
    warn "tenv was detected but install/use failed"
  fi

  warn "Could not auto-remediate Terraform with installed managers."
  warn "Manual options:"
  warn "  - mise install terraform@$target && mise use terraform@$target"
  warn "  - asdf install terraform $target && asdf set terraform $target"
  warn "  - tfenv install $target && tfenv use $target"
  warn "  - tenv tf install $target && tenv tf use $target"
  return 1
}

print_environment_context() {
  log "Validating local toolchain (mode: fix=$FIX_MODE strict=$STRICT_MODE)"
  if is_vscode_terminal; then
    warn "VS Code integrated terminal detected; inherited environment variables may be stale."
    warn "If paths/versions look incorrect after changes, restart the integrated terminal or reload VS Code window."
  fi

  if [[ -n "${IN_NIX_SHELL:-}" || -n "${NIX_PROFILES:-}" ]]; then
    warn "Nix shell/develop environment detected. This script will only report/advice and not mutate Nix configs."
  fi
}

main() {
  local failed=0

  print_environment_context

  check_required_command bash || failed=1
  check_required_command git || failed=1

  if ! validate_go_environment; then
    failed=1
    if (( FIX_MODE == 1 )); then
      attempt_fix_go || true
      log "Re-validating Go after remediation attempt"
      validate_go_environment || failed=1
    fi
  fi

  if ! validate_terraform; then
    failed=1
    if (( FIX_MODE == 1 )); then
      attempt_fix_terraform || true
      log "Re-validating Terraform after remediation attempt"
      validate_terraform || failed=1
    fi
  fi

  if (( failed == 0 )); then
    log "Toolchain validation complete: PASS"
    exit 0
  fi

  error "Toolchain validation complete: FAIL"
  if is_vscode_terminal; then
    warn "If you changed tool versions during this run, restart VS Code terminal and re-run ./scripts/setup-dev.sh"
  fi
  exit 1
}

main "$@"
