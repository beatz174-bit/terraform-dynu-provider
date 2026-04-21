#!/usr/bin/env bash
set -euo pipefail

echo "[test-integration] running mock-backed provider integration tests"
go test ./internal/provider -run '^TestIntegration' -count=1 -v
