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

if [[ "${DYNU_ACC:-}" != "1" ]]; then
  echo "[testacc][error] live mode requires DYNU_ACC=1" >&2
  exit 1
fi

if [[ -z "${DYNU_ACC_API_KEY:-}" ]]; then
  echo "[testacc][error] live mode requires DYNU_ACC_API_KEY" >&2
  exit 1
fi

if [[ -z "${DYNU_ACC_TEST_DOMAIN:-}" ]]; then
  echo "[testacc][error] live mode requires DYNU_ACC_TEST_DOMAIN" >&2
  exit 1
fi

echo "[testacc] running live acceptance tests (destructive: creates/updates/deletes records)"
TF_ACC=1 DYNU_API_KEY="${DYNU_ACC_API_KEY}" DYNU_DOMAIN="${DYNU_ACC_TEST_DOMAIN}" go test ./internal/provider -run '^TestAcc' -count=1 -v
