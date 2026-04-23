# Live Safe DNS Record Example

This example is an **opt-in live test** for writable Dynu provider functionality.
It creates exactly one disposable DNS `A` record under a user-supplied Dynu root domain,
then you remove it with `terraform destroy`.

## Safety model

- Creates a **new unique subdomain** every run using `random_id`.
- Hostname format is:
  - `<test_prefix>-<random_hex>.<dynu_root_domain>`
- Creates exactly one record:
  - `A` record, default content `198.51.100.10`, default TTL `300`.
- Does **not** import, update, replace, or delete existing production hostnames/records.
- `terraform destroy` removes only the disposable record in this state.

> [!WARNING]
> Do not repurpose this example to target existing production subdomains.
> Always provide only a root Dynu-managed domain (for example, `example.com`) via
> `dynu_root_domain`.

## Prerequisites

- Local provider binary and Terraform `dev_overrides` set up for `dynu/dynu`.
- `DYNU_API_KEY` exported (or set `dynu_api_key` in `terraform.tfvars`).
- A Dynu-managed root domain you control.

## Workflow

```bash
go build -o terraform-provider-dynu
cd examples/live_safe_dns_record
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars and set only dynu_root_domain
terraform validate
terraform plan
terraform apply
```

After apply, inspect outputs to confirm the generated disposable hostname and record ID,
then clean up:

```bash
terraform destroy
```

## Cleanup guidance

- Always run `terraform destroy` after testing.
- If apply is interrupted or partially succeeds, rerun `terraform destroy` from this same
  directory with the same `terraform.tfvars` and state files.
- If re-running without destroying first, Terraform may keep state for the previous
  disposable record until it is destroyed.
