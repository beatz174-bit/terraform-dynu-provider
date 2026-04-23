#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "Usage: ./build.sh <version>"
}

if [[ $# -ne 1 ]]; then
  usage
  exit 1
fi

if [[ ! -f "README.md" || ! -f "go.mod" || ! -f "main.go" ]]; then
  echo "Error: run this script from the repository root (terraform-dynu-provider)."
  exit 1
fi

if ! command -v git >/dev/null 2>&1; then
  echo "Error: git is required but was not found in PATH."
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "Error: go is required but was not found in PATH."
  exit 1
fi

VERSION="$1"
COMMIT="$(git rev-parse --short HEAD)"
DATE="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
OUTPUT="terraform-provider-dynu"

echo "Building terraform-provider-dynu with metadata:"
echo "  version: ${VERSION}"
echo "  commit:  ${COMMIT}"
echo "  date:    ${DATE}"
echo "  output:  ${OUTPUT}"

go build -o "${OUTPUT}" \
  -ldflags="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}"

echo "Build complete: ${OUTPUT}"
