# terraform-provider-dynu

Terraform provider for Dynu DNS using the Terraform Plugin Framework.

> Current phase is **read-only**: this provider implements provider configuration and data sources only.

## Features

- Provider configuration with API key authentication.
- Environment variable support (`DYNU_API_KEY`).
- Read-only data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`

## Requirements

- Terraform >= 1.5
- Go >= 1.22 (for building)
- A Dynu API key

## Provider configuration

```hcl
provider "dynu" {
  api_key = var.dynu_api_key
}
```

You can also omit `api_key` and set `DYNU_API_KEY` in your environment.

## Data sources

### dynu_domains

Returns all DNS domains associated with the account.

```hcl
data "dynu_domains" "all" {}
```

### dynu_domain

Resolves the root domain from a hostname, then returns full domain details.

```hcl
data "dynu_domain" "selected" {
  hostname = "www.example.com"
}
```

### dynu_dns_records

Resolves the root domain from a hostname, then returns DNS records for that domain.

```hcl
data "dynu_dns_records" "records" {
  hostname = "www.example.com"
}
```

## Development

```bash
./codex/setup.sh
./codex/maintain.sh
```

Run tests:

```bash
go test ./...
```

Acceptance tests (requires live Dynu credentials):

```bash
TF_ACC=1 DYNU_API_KEY=... go test ./internal/provider -run TestAcc
```

## Limitations

- No Terraform resources are implemented in this phase.
- No write operations are supported by provider code in this phase.
