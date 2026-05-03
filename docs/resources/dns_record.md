# dynu_dns_record Resource

Manages a Dynu DNS record under the Dynu root domain resolved from `hostname`.

## Example Usage

### A record

```terraform
resource "dynu_dns_record" "a" {
  hostname    = "www.example.com"
  record_type = "A"
  content     = "198.51.100.20"
  ttl         = 300
}
```

### CNAME record

```terraform
resource "dynu_dns_record" "cname" {
  hostname    = "app.example.com"
  record_type = "CNAME"
  content     = "target.example.net"
}
```

### MX record

```terraform
resource "dynu_dns_record" "mx" {
  hostname    = "example.com"
  record_type = "MX"
  content     = "mail.example.com"
  priority    = 10
}
```

### TXT record

```terraform
resource "dynu_dns_record" "txt" {
  hostname    = "example.com"
  record_type = "TXT"
  content     = "v=spf1 include:_spf.example.com ~all"
}
```

### SRV record

```terraform
resource "dynu_dns_record" "srv" {
  hostname    = "_sip._tcp.example.com"
  record_type = "SRV"
  content     = "sip.example.com"
  priority    = 10
  weight      = 5
  port        = 5060
}
```

## Schema

### Required

- `hostname` (String) Fully-qualified hostname.
- `record_type` (String) DNS record type.

### Optional + Computed

- `content` (String)
- `dynamic` (Boolean)
- `ttl` (Number)
- `enabled` (Boolean, defaults to `true`)
- `group` (String)
- `host` (String)
- `priority` (Number)
- `weight` (Number)
- `port` (Number)
- `flags` (Number)
- `tag` (String)
- `value` (String)
- `node_name` (String)

Type-specific notes:
- `A`/`AAAA` support dynamic semantics using omitted `content` or explicit `dynamic = true`.
- `MX` and `SRV` use `priority`.
- `SRV` also uses `weight` and `port`.
- `CAA` uses `flags`, `tag`, and `value`.

### Computed

- `id` (String) Composite ID in `domain_id/record_id` format.
- `domain_id` (Number)
- `domain_name` (String)
- `updated_on` (String)

## Import

Import using `domain_id/record_id`:

```bash
terraform import dynu_dns_record.example 1234/5678
```
