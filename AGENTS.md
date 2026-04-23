# AGENTS.md

## Goal
Build and maintain a standalone Terraform provider for Dynu DNS and domains.

## Scope
Current phase is **read-only**.

Implement:
- Provider configuration
- Environment variable support for credentials
- Read-only data sources:
  - `dynu_domains`
  - `dynu_domain`
  - `dynu_dns_records`

Do not implement in this phase:
- Any Terraform resources with create/update/delete
- Any write HTTP methods

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

## Portability requirements
- Keep the repository generic and standalone
- Do not reference personal infrastructure, private domains, homelab tooling, or external repo scripts
- Use neutral placeholders in docs and examples (for example: `example.com`, `my-test-domain.example`, `var.dynu_api_key`)
- Avoid hardcoded local filesystem paths

## Testing
- Add unit tests where practical
- Gate acceptance tests behind generic environment variables
  - `TF_ACC=1`
  - `DYNU_API_KEY`
  - optional `DYNU_DOMAIN` for domain-specific acceptance coverage
- Unit tests must not require live credentials

## Documentation
- Keep README and examples aligned with actual provider behavior
- Document current read-only limitations clearly
- Include build, test, and acceptance test instructions for any contributor

## Developer scripts
- `scripts/setup-dev.sh`: validate local toolchain requirements
- `scripts/check.sh`: run formatting and unit checks, then run Terraform checks against `examples/`
- `scripts/testacc.sh`: run acceptance tests using generic env vars

## Output expectations
Changes should keep the provider idiomatic, easy to contribute to, and ready for public/open-source usage.
