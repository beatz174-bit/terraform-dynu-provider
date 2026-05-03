# terraform-provider-dynu

A standalone Terraform provider for Dynu DNS and domain management.

## Features

- Full CRUD for Dynu root domains via `dynu_domain`.
- Full CRUD for Dynu DNS records via `dynu_dns_record`.
- Read-only discovery data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`
- Provider authentication via `api_key` or `DYNU_API_KEY`.

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

resource "dynu_domain" "example" {
  name = "my-test-domain.example"
  ttl  = 300
}

resource "dynu_dns_record" "www" {
  hostname    = "www.${dynu_domain.example.name}"
  record_type = "A"
  content     = "198.51.100.20"
  ttl         = 300
}
```

For a live end-to-end workflow that exercises multiple record types, see `examples/live_safe_dns_record/README.md`.

## Resources

- `dynu_domain`
- `dynu_dns_record`

## Data sources

- `dynu_domains`
- `dynu_domain`
- `dynu_dns_records`

## Testing

Run Go unit/integration tests:

```bash
go test ./...
```

Run repository checks:

```bash
./scripts/fix.sh
./scripts/check.sh
```

Run the live end-to-end Terraform example (opt-in, uses real Dynu account data):

```bash
cd examples/live_safe_dns_record
cp terraform.tfvars.example terraform.tfvars
terraform validate
terraform plan
# terraform apply
# terraform destroy
```

## Development

This provider is not yet published to the Terraform Registry. Use `dev_overrides` with a local build.

1. Build the provider binary:

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

3. Validate locally without relying on registry publishing:

```bash
cd examples/read_only
cp terraform.tfvars.example terraform.tfvars
terraform validate
terraform plan
```

When provider code/config changes, rebuild `terraform-provider-dynu` before re-running Terraform commands.
