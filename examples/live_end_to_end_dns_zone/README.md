# Live End-to-End Dynu DNS Zone Example

> [!WARNING]
> **This is a live destructive test.**
>
> - Do **not** use a real production domain.
> - Use only a disposable test domain.
> - `terraform destroy` will remove the test domain and all records created by this example.

This example performs full live validation of the Dynu provider lifecycle by creating and then destroying a disposable Dynu domain plus one DNS record for each currently supported record type in this provider example:

- A
- AAAA
- CNAME
- MX
- TXT
- SRV
- CAA

## Prerequisites

- A local build of this provider with Terraform `dev_overrides` for `beatz174-bit/dynu`.
- Dynu API key provided as Terraform variable `dynu_api_key` (for example via `terraform.tfvars`).
- A disposable domain value for `test_domain`.

## Required variables

Create `terraform.tfvars` in this folder:

```hcl
test_domain = "my-disposable-test-domain.example"
dynu_api_key = "..."
```

## Commands

```bash
terraform init
terraform validate
terraform plan
terraform apply
terraform plan
terraform destroy
```

## Expected results

- `terraform validate` should pass.
- First `terraform plan` should show creation of:
  - one disposable test domain (`dynu_domain.test`), and
  - one record each for A, AAAA, CNAME, MX, TXT, SRV, and CAA.
- `terraform apply` should create the domain and records.
- Second `terraform plan` (after apply) should show **no changes**.
- Manually verify in the Dynu web UI that the test domain and records exist after apply.
- `terraform destroy` should remove the test domain and all records from state.
- Final manual verification should confirm the test domain is gone from Dynu.
