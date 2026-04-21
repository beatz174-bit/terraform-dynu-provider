---
page_title: "dynu_domain Data Source"
description: |-
  Resolves a hostname to its Dynu root domain and returns domain details.
---

# dynu_domain Data Source

## Example Usage

```terraform
data "dynu_domain" "selected" {
  hostname = "www.example.com"
}
```

## Schema

### Required

- `hostname` (String) Fully-qualified hostname.

### Read-Only

- `domain` (Object) Same fields returned by `dynu_domains.domains[*]`.
