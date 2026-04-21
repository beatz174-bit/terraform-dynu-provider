---
page_title: "dynu_dns_records Data Source"
description: |-
  Lists DNS records for the Dynu root domain resolved from a hostname.
---

# dynu_dns_records Data Source

## Example Usage

```terraform
data "dynu_dns_records" "records" {
  hostname = "www.example.com"
}
```

## Schema

### Required

- `hostname` (String) Fully-qualified hostname.

### Read-Only

- `domain_id` (Number)
- `domain_name` (String)
- `records` (List of Object)
  - `id` (Number)
  - `domain_id` (Number)
  - `domain_name` (String)
  - `node_name` (String)
  - `hostname` (String)
  - `record_type` (String)
  - `ttl` (Number)
  - `state` (Boolean)
  - `content` (String)
  - `updated_on` (String)
  - `group` (String)
  - `host` (String)
