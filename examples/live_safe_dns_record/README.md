# Live Safe DNS Record Example

This example is an **opt-in live test** for writable Dynu provider functionality.
It creates exactly one disposable DNS `A` record under a user-supplied Dynu root domain,
then you remove it with `terraform destroy`.

## Safety model

- Creates a **new disposable subdomain** using configurable `test_prefix` and `test_suffix` values.
- Hostname format is:
  - `<test_prefix>-<test_suffix>.<dynu_root_domain>`
- Creates exactly one record:
  - `A` record, default content `1.1.1.1`, default TTL `300`.
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

This example intentionally uses only the `dynu/dynu` provider (no `hashicorp/random`).
If you want per-run uniqueness, pass `test_suffix` explicitly.

```bash
go build -o terraform-provider-dynu
cd examples/live_safe_dns_record
cp terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars and set dynu_root_domain (and optionally test_suffix)
terraform validate
TEST_SUFFIX="$(date +%s)"
terraform plan -var="test_suffix=${TEST_SUFFIX}"
terraform apply -var="test_suffix=${TEST_SUFFIX}"
```

After apply, inspect outputs to confirm the disposable hostname and record ID,
then clean up using the **same** suffix value:

```bash
terraform destroy -var="test_suffix=${TEST_SUFFIX}"
```

## Cleanup guidance

- Always run `terraform destroy` after testing.
- If apply is interrupted or partially succeeds, rerun `terraform destroy` from this same
  directory with the same `terraform.tfvars`, CLI vars, and state files.
- If re-running without destroying first, Terraform may keep state for the previous
  disposable record until it is destroyed.
