# Live Safe DNS Record Example

This example is an **opt-in live test** for writable Dynu provider functionality. It creates real Dynu DNS records and is designed for disposable, test-owned hostnames only.

## What this creates

Using a single suffix (`test_suffix`), this example creates five DNS record scenarios under `dynu_root_domain`:

1. `A` record with IPv4 content (`codex-a-<suffix>.<root_domain>`)
2. `AAAA` record with IPv6 content (`codex-aaaa-<suffix>.<root_domain>`)
3. `CNAME` record (`codex-cname-<suffix>.<root_domain>`)
4. **Dynamic `A` record** with omitted content (`codex-dynamic-a-<suffix>.<root_domain>`)
5. **Dynamic `AAAA` record** with omitted content (`codex-dynamic-aaaa-<suffix>.<root_domain>`)

> [!WARNING]
> Do not use a suffix that overlaps important existing hostnames. This example is intended only for disposable test records that you can safely destroy.

## Prerequisites

- Local provider binary + Terraform `dev_overrides` for `dynu/dynu`
- `DYNU_API_KEY` exported, or `dynu_api_key` set in `terraform.tfvars`
- A Dynu-managed root domain you control

## Configure

```bash
cp terraform.tfvars.example terraform.tfvars
```

Edit `terraform.tfvars` and set at least:

- `dynu_root_domain`
- `test_suffix` (use a unique value per run)

Optional overrides include `test_ipv4`, `test_ipv6`, and `test_cname_target`. Use real routable IPs for `test_ipv4`/`test_ipv6`; Dynu rejects documentation ranges such as `192.0.2.0/24` and `2001:db8::/32`.

## Run

```bash
terraform init
terraform validate
terraform plan
terraform apply
terraform destroy
```

## Notes

- `terraform apply` should create all five record scenarios.
- `terraform destroy` should remove all five records created by this state.
- If you need to target a single scenario, resources are explicitly named:
  - `dynu_dns_record.a_ipv4`
  - `dynu_dns_record.aaaa_ipv6`
  - `dynu_dns_record.cname`
  - `dynu_dns_record.dynamic_a`
  - `dynu_dns_record.dynamic_aaaa`

Verification commands:

```bash
dig A codex-dynamic-a-<suffix>.<root_domain>
dig AAAA codex-dynamic-aaaa-<suffix>.<root_domain>
dig CNAME codex-cname-<suffix>.<root_domain>
```

`dig <hostname>` defaults to an `A` lookup, which can be misleading for AAAA-only checks.
