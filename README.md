# terraform-provider-dynu

A standalone Terraform provider for Dynu DNS.

> Status: **read-only milestone**. This provider currently implements provider configuration and data sources only.

## Feature scope

Implemented:
- Provider authentication using `api_key` or `DYNU_API_KEY`
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
- Dynu API key

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
- `./scripts/check.sh` - formatting, vet, and unit tests
- `./scripts/testacc.sh` - acceptance tests only

### Standalone repository guarantee

This repository is intentionally self-contained:
- no dependency on sibling repositories
- no dependency on external helper scripts (for example `services-up.sh`)
- no hardcoded local paths (for example `/workspace/...` or `/home/...`)

### Build

```bash
go build ./...
```

### Unit tests

```bash
go test ./...
```

### Acceptance tests

Acceptance tests are read-only and opt-in.

Required environment variables:
- `TF_ACC=1`
- `DYNU_API_KEY`

Optional:
- `DYNU_DOMAIN` (required for domain-specific acceptance tests such as `dynu_domain` and `dynu_dns_records`)

Run:

```bash
TF_ACC=1 DYNU_API_KEY="your-dynu-api-key" DYNU_DOMAIN="www.example.com" ./scripts/testacc.sh
```

If `DYNU_DOMAIN` is omitted, domain-specific tests skip cleanly.

## CI

GitHub Actions CI runs on push and pull requests and executes:
- gofmt verification
- `go vet ./...`
- `go test ./...`

Acceptance tests are intentionally excluded from default CI.

## Documentation

Registry-style markdown docs are stored in `docs/`.

## Limitations

- Dynu timestamps are currently exposed as strings exactly as returned by Dynu API.
- Data returned from Dynu is sorted in provider state for Terraform stability.
- Read-only operations only.

## Roadmap

Next planned milestone after this quality-hardening release:
- first writable resource (`dynu_dns_record`) with careful CRUD behavior and acceptance coverage.
