---
page_title: "dynu_domains Data Source"
description: |-
  Lists domains visible to the configured Dynu API key.
---

# dynu_domains Data Source

## Example Usage

```terraform
data "dynu_domains" "all" {}
```

## Read-Only Attributes

- `domains` (List of Object)
  - `id` (Number)
  - `name` (String)
  - `unicode_name` (String)
  - `token` (String, Sensitive)
  - `state` (String)
  - `group` (String)
  - `ipv4_address` (String)
  - `ipv6_address` (String)
  - `ttl` (Number)
  - `ipv4` (Boolean)
  - `ipv6` (Boolean)
  - `ipv4_wildcard_alias` (Boolean)
  - `ipv6_wildcard_alias` (Boolean)
  - `allow_zone_transfer` (Boolean)
  - `dnssec` (Boolean)
  - `created_on` (String)
  - `updated_on` (String)
