# terraform-provider-dynu

A standalone Terraform provider for [Dynu](https://www.dynu.com/) DNS data.

> Current phase: **read-only**. This provider supports provider configuration and data sources only.

## Feature scope

Implemented:
- Provider authentication via API key
- Environment variable support (`DYNU_API_KEY`)
- Read-only data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`

Not implemented in this phase:
- Terraform resources
- Any create/update/delete operations

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) `>= 1.5`
- [Go](https://go.dev/dl/) `>= 1.23` (for building and testing)
- Dynu API key

## Provider configuration

```hcl
provider "dynu" {
  api_key = var.dynu_api_key
}

variable "dynu_api_key" {
  type      = string
  sensitive = true
}
```

You can omit `api_key` in Terraform configuration and set the environment variable instead:

```bash
export DYNU_API_KEY="your-dynu-api-key"
```

## Data source usage

### dynu_domains

```hcl
data "dynu_domains" "all" {}
```

### dynu_domain

```hcl
data "dynu_domain" "selected" {
  hostname = "www.example.com"
}
```

### dynu_dns_records

```hcl
data "dynu_dns_records" "records" {
  hostname = "www.example.com"
}
```

## Build

```bash
go build ./...
```

## Test

Run formatting and unit tests:

```bash
./scripts/check.sh
```

Or run unit tests directly:

```bash
go test ./...
```

## Acceptance tests

Acceptance tests are opt-in and require live Dynu credentials.

```bash
TF_ACC=1 DYNU_API_KEY="your-dynu-api-key" ./scripts/testacc.sh
```

Optional environment variable:
- `DYNU_DOMAIN` (for future domain-specific acceptance test cases)

## Developer workflow

- `./scripts/setup-dev.sh` – validates required local tools
- `./scripts/check.sh` – runs formatting and unit checks
- `./scripts/testacc.sh` – runs acceptance tests

Repository-local Codex helpers are also available under `codex/` for agent-oriented workflows.

## Limitations

- Read-only provider phase only
- No writable Terraform resources yet
