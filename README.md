# terraform-provider-dynu

A standalone Terraform provider for Dynu DNS.

> Status: **read-only milestone**. This provider currently implements provider configuration plus read-only data sources.

## Quick start (local dev with `dev_overrides`)

This provider is **not published** to the Terraform Registry yet. For local development, use Terraform CLI `dev_overrides` and your local provider binary.

1. Build provider binary in repo root:

```bash
go build -o terraform-provider-dynu
```

2. Configure `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "dynu/dynu" = "/path/to/terraform-dynu-provider"
  }

  direct {}
}
```

3. Run the runnable read-only example:

```bash
cd examples/read_only
cp terraform.tfvars.example terraform.tfvars
terraform validate
terraform plan
```

> With `dev_overrides`, Terraform uses your local binary for `dynu/dynu`. `terraform init` is not the primary local-dev loop here and may still try registry/network operations.

## Copy/paste starter configuration

```hcl
terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {
  api_key = var.dynu_api_key # optional if DYNU_API_KEY is set
}

variable "dynu_api_key" {
  type      = string
  default   = null
  sensitive = true
}

data "dynu_domains" "all" {}

# Use a real hostname from your Dynu account.
data "dynu_domain" "selected" {
  hostname = "www.example.com"
}

data "dynu_dns_records" "selected" {
  hostname = "www.example.com"
}
```

## Provider schema reference

### Provider: `dynu`

Optional arguments:
- `api_key` (String, Sensitive)
  - Falls back to `DYNU_API_KEY` environment variable when omitted.
- `base_url` (String)
  - Test/dev override for Dynu API base URL.

No provider resources are implemented yet.

## Data source schema reference

### `dynu_domains`

Arguments:
- none

Attributes:
- `domains` (List(Object)):
  - `id`, `name`, `unicode_name`, `token` (sensitive), `state`, `group`
  - `ipv4_address`, `ipv6_address`, `ttl`
  - `ipv4`, `ipv6`, `ipv4_wildcard_alias`, `ipv6_wildcard_alias`
  - `allow_zone_transfer`, `dnssec`, `created_on`, `updated_on`

Example:

```hcl
data "dynu_domains" "all" {}
```

### `dynu_domain`

Arguments:
- `hostname` (String, required)

Attributes:
- `domain` (Object) with the same fields as `dynu_domains.domains[*]`.

Example:

```hcl
data "dynu_domain" "selected" {
  hostname = "www.example.com"
}
```

### `dynu_dns_records`

Arguments:
- `hostname` (String, required)

Attributes:
- `domain_id` (Number)
- `domain_name` (String)
- `records` (List(Object)) with:
  - `id`, `domain_id`, `domain_name`, `node_name`, `hostname`, `record_type`
  - `ttl`, `state`, `content`, `updated_on`, `group`, `host`

Example:

```hcl
data "dynu_dns_records" "selected" {
  hostname = "www.example.com"
}
```

## Examples

- Runnable local workflow: `examples/read_only/`
- Provider block example: `examples/provider/provider.tf`
- Individual data source snippets:
  - `examples/data-sources/dynu_domains/data-source.tf`
  - `examples/data-sources/dynu_domain/data-source.tf`
  - `examples/data-sources/dynu_dns_records/data-source.tf`

## Troubleshooting local dev

- **Unsupported provider arguments**
  - Symptom: errors such as `Unsupported argument` (for example `username`).
  - Fix: use only `api_key` and/or `base_url` in `provider "dynu"`.

- **Bad API credentials**
  - Symptom: diagnostics mention authentication failures.
  - Fix: verify `api_key` or `DYNU_API_KEY` and re-run `terraform plan`.

- **Unknown data source arguments**
  - Symptom: unsupported argument errors in data blocks.
  - Fix: `dynu_domain` and `dynu_dns_records` require only `hostname`; `dynu_domains` takes no arguments.

- **Stale provider binary after code changes**
  - Symptom: Terraform behavior doesn't reflect latest code.
  - Fix: rebuild binary (`go build -o terraform-provider-dynu`) and run `terraform plan` again.

## Developer workflow

- `./scripts/setup-dev.sh` - validate local toolchain requirements
- `./scripts/check.sh` - formatting, vet, and unit tests
- `./scripts/test-integration.sh` - local mock-backed provider integration tests
- `./scripts/testacc.sh` - acceptance/integration test wrapper (live tests opt-in)

Live acceptance tests are read-only and require:
- `TF_ACC=1`
- `DYNU_API_KEY`
- optional `DYNU_DOMAIN` for domain-specific coverage

## Feature scope

Implemented:
- Provider authentication via `api_key` or `DYNU_API_KEY`
- Optional provider `base_url` override
- Data sources: `dynu_domains`, `dynu_domain`, `dynu_dns_records`

Not implemented yet:
- Terraform resources (create/update/delete)
- Any write API operations
