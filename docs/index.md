---
page_title: "dynu Provider"
description: |-
  Terraform provider for Dynu DNS read-only data sources.
---

# dynu Provider

The `dynu` provider lets Terraform read Dynu DNS domain and record data.
This provider repository is standalone and does not require any external companion repository or helper script.

## Example Usage

```terraform
terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {
  api_key = var.dynu_api_key
}
```

## Schema

### Optional

- `api_key` (String, Sensitive) Dynu API key. If omitted, `DYNU_API_KEY` is used.
- `base_url` (String) Override API base URL. Primarily intended for automated tests.
