# terraform-provider-dynu

A standalone Terraform provider for Dynu DNS and domain management.

## Features

- Full CRUD for Dynu root domains via `dynu_domain`.
- Full CRUD for Dynu DNS records via `dynu_dns_record`.
- Read-only discovery data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`
- Provider authentication via explicit `api_key` configuration.

## Important safety note

Deleting `dynu_domain` deletes the full Dynu DNS zone for that domain. Treat destroy plans carefully.

## Minimal usage example

```hcl
terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {
  api_key = var.dynu_api_key
}

variable "dynu_api_key" {
  type      = string
  sensitive = true
}
```

## Resources

- `dynu_domain`
- `dynu_dns_record`

## Data sources

- `dynu_domains`
- `dynu_domain`
- `dynu_dns_records`

## Local development and dev overrides

This provider is not yet published to the Terraform Registry. Use `dev_overrides` with a local build.

```bash
go build -o terraform-provider-dynu
```

`~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "dynu/dynu" = "/path/to/terraform-dynu-provider"
  }

  direct {}
}
```

Validate locally:

```bash
cd examples/read_only
cp terraform.tfvars.example terraform.tfvars
terraform validate
terraform plan
```

If provider code/config changes, rebuild `terraform-provider-dynu` before running Terraform again.

## Testing

```bash
./scripts/fix.sh
./scripts/check.sh
go test ./...
go vet ./...
terraform fmt -check -recursive examples
```

### Optional live acceptance tests

Live tests are opt-in and destructive for test records. They never run by default.

```bash
DYNU_ACC=1 \
DYNU_ACC_API_KEY="***" \
DYNU_ACC_TEST_DOMAIN="example.com" \
./scripts/testacc.sh --live
```

Use a disposable domain/subdomain only.

## Release

Build a local stamped binary:

```bash
./build.sh v0.1.0
```

Tagged releases (`v*`) run `.github/workflows/release.yml` with GoReleaser to produce multi-platform archives and checksums.

Terraform Registry signing is not configured in this repository yet.
