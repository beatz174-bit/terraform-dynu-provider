#!/usr/bin/env bash
set -euo pipefail

log() {
  printf '[codex] %s\n' "$*"
}

warn() {
  printf '[codex][warn] %s\n' "$*" >&2
}

die() {
  printf '[codex][error] %s\n' "$*" >&2
  exit 1
}

have() {
  command -v "$1" >/dev/null 2>&1
}

repo_root() {
  git rev-parse --show-toplevel 2>/dev/null || pwd
}

cd_repo_root() {
  cd "$(repo_root)"
}

ensure_file() {
  local path="$1"
  [[ -f "$path" ]] || die "Required file not found: $path"
}

ensure_dir() {
  local path="$1"
  mkdir -p "$path"
}

append_if_missing() {
  local line="$1"
  local file="$2"
  touch "$file"
  grep -Fqx "$line" "$file" || printf '%s\n' "$line" >> "$file"
}
