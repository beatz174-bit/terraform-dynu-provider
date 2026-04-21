# terraform-provider-dynu

A standalone Terraform provider for Dynu DNS.

> Status: **read-only milestone**. This provider currently implements provider configuration and data sources only.

## Feature scope

Implemented:
- Provider authentication using `api_key` or `DYNU_API_KEY`
- Optional provider `base_url` override for local test/dev setups
- Data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`

Not implemented yet:
- Terraform resources (no create/update/delete)
- Any write API operations

## Provider source and module path

- Terraform provider source address: `dynu/dynu`
- Go module path: `github.com/dynu/terraform-provider-dynu`

The repository can be hosted elsewhere during development, but module and provider source naming are kept aligned with planned public registry publishing.

## Requirements

- Terraform `>= 1.5`
- Go `>= 1.23`
- Dynu API key for live API usage

## Authentication

Option 1: Terraform configuration.

```hcl
provider "dynu" {
  api_key = var.dynu_api_key
}

variable "dynu_api_key" {
  type      = string
  sensitive = true
}
```

Option 2: Environment variable.

```bash
export DYNU_API_KEY="your-dynu-api-key"
```

## Data source examples

See the `examples/` directory:
- `examples/provider/provider.tf`
- `examples/data-sources/dynu_domains/data-source.tf`
- `examples/data-sources/dynu_domain/data-source.tf`
- `examples/data-sources/dynu_dns_records/data-source.tf`

## Developer workflow

- `./scripts/setup-dev.sh` - verify required local tools
- `./scripts/check.sh` - formatting, vet, and unit tests (Tier A)
- `./scripts/test-integration.sh` - local mock-backed provider integration tests (Tier B)
- `./scripts/testacc.sh` - default: Tier B; live mode available with `--live` (Tier C)

### Standalone repository guarantee

This repository is intentionally self-contained:
- no dependency on sibling repositories
- no dependency on external helper scripts (for example `services-up.sh`)
- no hardcoded local paths (for example `/workspace/...` or `/home/...`)

### Build

```bash
go build ./...
```

## Testing model

The provider now has three explicit test tiers:

### Tier A: unit tests (fast, no network)

Covers focused package behavior (client parsing, mappers, provider helper logic).

```bash
./scripts/check.sh
go test ./...
```

### Tier B: local integration tests (mock Dynu API, no real credentials)

These tests use an `httptest` fake Dynu API server and run the Terraform provider end-to-end against deterministic fixtures.

- No Dynu account required
- Dummy API key is used in test provider configuration
- Exercises provider wiring, schema/state mapping, hostname resolution flow, and diagnostic behavior

```bash
./scripts/test-integration.sh
./scripts/testacc.sh
```

### Tier C: live acceptance tests (opt-in)

These tests call the real Dynu API and are read-only.

Required environment variables:
- `TF_ACC=1`
- `DYNU_API_KEY`

Optional:
- `DYNU_DOMAIN` (required for domain-specific acceptance tests such as `dynu_domain` and `dynu_dns_records`)

```bash
TF_ACC=1 DYNU_API_KEY="your-dynu-api-key" DYNU_DOMAIN="www.example.com" ./scripts/testacc.sh --live
# or
LIVE=1 TF_ACC=1 DYNU_API_KEY="your-dynu-api-key" ./scripts/testacc.sh
```

If `DYNU_DOMAIN` is omitted, domain-specific live tests skip cleanly.

## CI

GitHub Actions CI runs on push and pull requests and executes:
- gofmt verification
- `go vet ./...`
- `go test ./...`

Live acceptance tests are intentionally excluded from default CI.

## Documentation

Registry-style markdown docs are stored in `docs/`.

## Limitations

- Dynu timestamps are currently exposed as strings exactly as returned by Dynu API.
- Data returned from Dynu is sorted in provider state for Terraform stability.
- Read-only operations only.

## Roadmap

Next planned milestone after this testing foundation:
- first writable resource (`dynu_dns_record`) with strict schema validation, import support, mock-first integration tests, and then live acceptance coverage.
