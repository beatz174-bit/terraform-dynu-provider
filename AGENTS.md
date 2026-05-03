# AGENTS.md

## Goal
Build and maintain a standalone Terraform provider for Dynu DNS and domains.

## Scope
This provider supports full CRUD for Dynu DNS domains and DNS records.

Current inventory:

### Resources
- `dynu_domain` (CRUD + import)
- `dynu_dns_record` (CRUD + import)

### Data sources
- `dynu_domains`
- `dynu_domain`
- `dynu_dns_records`

## Language and framework
- Language: Go
- Framework: HashiCorp Terraform Plugin Framework
- Do not use the legacy Terraform Plugin SDK unless explicitly requested

## API usage
- Use Dynu public API endpoints
- Prefer stable, documented endpoints
- Keep HTTP logic in a small internal client package
- Handle pagination and API errors clearly
- Never log secrets or credentials

## Implementation expectations for agents
When making changes, agents must:
- preserve backwards compatibility where possible
- ensure Terraform idempotency (no diff after apply)
- validate all API interactions against Dynu behaviour
- maintain consistent schema design (`optional` vs `computed` vs `required`)
- ensure tests pass before adding new features

## DNS record type guidance
- Multiple DNS record types are supported, including `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `SRV`, and `CAA` fields exposed by schema.
- Apply type-specific handling consistently:
  - `MX`/`SRV`: ensure priority semantics are correct.
  - `SRV`: ensure `weight` and `port` handling is preserved.
  - `TXT`: preserve exact value semantics and quoting expectations from Terraform configuration.
  - `CAA`: preserve `flags`, `tag`, and `value` field behavior.
- For dynamic `A`/`AAAA` behavior, avoid introducing drift-prone behavior or unsafe payload normalization.

## Portability requirements
- Keep the repository generic and standalone
- Do not reference personal infrastructure, private domains, homelab tooling, or external repo scripts
- Use neutral placeholders in docs and examples (for example: `example.com`, `my-test-domain.example`, `var.dynu_api_key`)
- Avoid hardcoded local filesystem paths

## Testing
- `go test ./...` must pass.
- Add unit tests where practical; new features must include tests.
- Gate acceptance tests behind generic environment variables:
  - `TF_ACC=1`
  - `DYNU_API_KEY`
  - optional `DYNU_DOMAIN` for domain-specific acceptance coverage
- Unit tests must not require live credentials.
- End-to-end Terraform example(s) must remain valid.

## Destructive operation warning
- Deleting `dynu_domain` removes the entire Dynu DNS zone for that domain.
- Agents must not introduce unsafe defaults or implicit destructive behavior.

## Documentation
- Keep README and examples aligned with actual provider behavior
- Include build, test, and acceptance test instructions for any contributor

## Developer scripts
- `scripts/setup-dev.sh`: validate local toolchain requirements
- `scripts/check.sh`: run formatting and unit checks, then run Terraform checks against `examples/`
- `scripts/testacc.sh`: run acceptance tests using generic env vars

## Output expectations
Changes should keep the provider idiomatic, easy to contribute to, and ready for public/open-source usage.

## Validation workflow (Codex and local)

When you modify **Go** or **Terraform** files, always run this sequence before concluding work:

1. `./scripts/fix.sh`
2. `./scripts/check.sh`

### Expectations
- Do not leave formatting changes unapplied.
- Do not hand-edit formatting that `gofmt` or `terraform fmt` can apply automatically.
- Treat `./scripts/check.sh` as the required final validation gate.
- If Terraform is unavailable in the environment, note that Terraform formatting/checking could not be run locally.

### Local validation for this repository

This provider is not yet published to the Terraform Registry.

For local development, Terraform must use the locally built provider binary via `dev_overrides`. Because of that, `terraform init` is **not** part of the normal validation loop for agents working in this repository.

Use this loop instead:

```bash
go build -o terraform-provider-dynu
cd examples/read_only
terraform validate
terraform plan
```

If provider configuration or code changes, rebuild the binary before running Terraform again.

Do not infer that `terraform init` is required just because it is common in normal Terraform projects.

Leave GitHub Actions, CI workflows, and CI test behavior unchanged for this documentation guidance.
