# AGENTS.md

## Goal
Build a Terraform provider for Dynu DNS and domains.

## Scope for current phase
Read-only only. Implement provider configuration and data sources first.
Do not implement any writable Terraform resources in the first phase.

## Language and framework
- Language: Go
- Terraform provider framework: HashiCorp Terraform Plugin Framework
- Do not use the legacy SDK unless explicitly requested.

## API usage
- Use Dynu's public API.
- Prefer stable, documented endpoints.
- Build a small internal API client package rather than scattering HTTP calls across resources/data sources.
- Add clear handling for pagination, 4xx/5xx responses, and malformed responses.
- Never log secrets.

## Provider design
Implement:
- Provider configuration
- Environment variable support for credentials
- Read-only data sources:
  - dynu_domains
  - dynu_domain
  - dynu_dns_records

Do not implement:
- resource_dynu_dns_record
- any create/update/delete flows
- any speculative unsupported resources

## Safety
- Read-only phase only.
- No write HTTP methods in this phase.
- Acceptance tests must only cover provider auth and read-only data sources.
- Any future write support must be added in a separate phase.

## Code quality
- Keep code modular and idiomatic Go.
- Use strong typing for API models.
- Prefer explicit schema definitions with clear descriptions.
- Return actionable diagnostics.
- Keep public documentation accurate to actual implementation.

## Testing
- Add unit tests where practical.
- Add acceptance tests gated behind environment variables.
- Do not require live credentials for normal unit tests.

## Documentation
- Add examples for every implemented data source.
- Update README with provider configuration and environment variables.
- Document limitations and unsupported areas clearly.

## Output expectations
Create a provider skeleton that can compile and expose the provider plus read-only data sources.

## Codex environment
- Use `codex/setup.sh` to prepare the repository-local Codex environment.
- Use `codex/maintain.sh` to refresh and validate the environment.
- Use `codex/doctor.sh` for troubleshooting.
- Do not install global packages unless explicitly required.
- Prefer repo-local state under `.codex/`.
