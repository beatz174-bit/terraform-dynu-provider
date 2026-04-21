#!/usr/bin/env bash
set -euo pipefail

mode="mock"
if [[ "${1:-}" == "--live" || "${LIVE:-}" == "1" ]]; then
  mode="live"
fi

if [[ "${mode}" == "mock" ]]; then
  echo "[testacc] no --live flag detected; running local mock-backed integration tests"
  exec "$(dirname "$0")/test-integration.sh"
fi

if [[ "${TF_ACC:-}" != "1" ]]; then
  echo "[testacc][error] live mode requires TF_ACC=1" >&2
  exit 1
fi

if [[ -z "${DYNU_API_KEY:-}" ]]; then
  echo "[testacc][error] live mode requires DYNU_API_KEY" >&2
  exit 1
fi

if [[ -z "${DYNU_DOMAIN:-}" ]]; then
  echo "[testacc][warn] DYNU_DOMAIN not set; domain-specific acceptance tests will skip"
fi

echo "[testacc] running live acceptance tests"
go test ./internal/provider -run '^TestAcc' -count=1 -v
