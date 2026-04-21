#!/usr/bin/env bash
set -euo pipefail

if [[ "${TF_ACC:-}" != "1" ]]; then
  echo "[testacc][error] TF_ACC must be set to 1" >&2
  exit 1
fi

if [[ -z "${DYNU_API_KEY:-}" ]]; then
  echo "[testacc][error] DYNU_API_KEY must be set" >&2
  exit 1
fi

if [[ -z "${DYNU_DOMAIN:-}" ]]; then
  echo "[testacc][warn] DYNU_DOMAIN not set; domain-specific acceptance tests will skip"
fi

echo "[testacc] running acceptance tests"
go test ./internal/provider -run '^TestAcc' -count=1 -v
