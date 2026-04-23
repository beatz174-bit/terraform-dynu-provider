# Live-Safe DNS Records Lookup Example

This example is an **opt-in live read-only test**.
It queries DNS records for an existing Dynu-managed hostname and performs no writes.

## Safety model

- Uses only the `dynu_dns_records` data source.
- Makes read-only API requests.
- Does **not** create, update, or delete any DNS records.

## Prerequisites

- Local provider binary and Terraform `dev_overrides` set up for `dynu/dynu`.
- `DYNU_API_KEY` exported (or set `dynu_api_key` in `terraform.tfvars`).
- An existing Dynu-managed hostname you control.

## Workflow

```bash
go build -o terraform-provider-dynu
cd examples/live_safe_dns_record
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars and set hostname
terraform validate
terraform plan
```

After planning, inspect outputs to review returned DNS records.
