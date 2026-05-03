# dynu_domain Resource

Manages a Dynu DNS domain.

## Example Usage

```terraform
resource "dynu_domain" "example" {
  name         = "my-test-domain.example"
  ipv4_address = "203.0.113.10"
  ttl          = 300
}

resource "dynu_dns_record" "www" {
  hostname    = "www.${dynu_domain.example.name}"
  record_type = "A"
  content     = "203.0.113.20"
}
```

## Import

Import using the numeric Dynu domain ID:

```bash
terraform import dynu_domain.example 1234
```

## Attributes

- `id` (Number) Dynu domain ID.
- `name` (String) Domain name (forces replacement when changed).
- `ipv4_address` (String) Optional IPv4 address.
- `ipv6_address` (String) Optional IPv6 address.
- `ttl` (Number) Optional TTL.
- `group` (String) Optional group.
- `state` (String) Computed state from Dynu API.
- `token` (String, Sensitive) Computed domain token from Dynu API.

## DNS record examples under this domain

```terraform
resource "dynu_dns_record" "mx" {
  hostname    = dynu_domain.example.name
  record_type = "MX"
  content     = "mail.my-test-domain.example"
  priority    = 10
}

resource "dynu_dns_record" "txt_spf" {
  hostname    = dynu_domain.example.name
  record_type = "TXT"
  content     = "v=spf1 include:_spf.example.com ~all"
}

resource "dynu_dns_record" "cname" {
  hostname    = "app.${dynu_domain.example.name}"
  record_type = "CNAME"
  content     = "target.example.net"
}

resource "dynu_dns_record" "srv" {
  hostname    = "_sip._tcp.${dynu_domain.example.name}"
  record_type = "SRV"
  content     = "sip.my-test-domain.example"
  priority    = 10
  weight      = 5
  port        = 5060
}
```
