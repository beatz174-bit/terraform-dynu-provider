# Local development note:
# Keep source as "dynu/dynu" and use ~/.terraformrc dev_overrides to point to
# your local terraform-provider-dynu binary during provider development.
provider "dynu" {
  api_key = var.dynu_api_key
}

locals {
  hostname_a_ipv4       = "codex-a-${var.test_suffix}.${var.dynu_root_domain}"
  hostname_aaaa_ipv6    = "codex-aaaa-${var.test_suffix}.${var.dynu_root_domain}"
  hostname_cname        = "codex-cname-${var.test_suffix}.${var.dynu_root_domain}"
  hostname_dynamic_a    = "codex-dynamic-a-${var.test_suffix}.${var.dynu_root_domain}"
  hostname_dynamic_aaaa = "codex-dynamic-aaaa-${var.test_suffix}.${var.dynu_root_domain}"
}

resource "dynu_dns_record" "a_ipv4" {
  hostname    = local.hostname_a_ipv4
  record_type = "A"
  content     = var.test_ipv4
  ttl         = 90
  enabled     = true
}

resource "dynu_dns_record" "aaaa_ipv6" {
  hostname    = local.hostname_aaaa_ipv6
  record_type = "AAAA"
  content     = var.test_ipv6
  ttl         = var.test_ttl
  enabled     = true
}

resource "dynu_dns_record" "cname" {
  hostname    = local.hostname_cname
  record_type = "CNAME"
  content     = var.test_cname_target
  ttl         = var.test_ttl
  enabled     = false
}

# Deliberate dynamic A record scenario: content intentionally omitted for Dynu dynamic IPv4 behavior.
resource "dynu_dns_record" "dynamic_a" {
  hostname    = local.hostname_dynamic_a
  record_type = "A"
  ttl         = var.test_ttl
  enabled     = true
}

# Deliberate dynamic AAAA record scenario: content intentionally omitted for Dynu dynamic IPv6 behavior.
resource "dynu_dns_record" "dynamic_aaaa" {
  hostname    = local.hostname_dynamic_aaaa
  record_type = "AAAA"
  ttl         = var.test_ttl
  enabled     = true
}
