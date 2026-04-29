resource "dynu_dns_record" "a_enabled_min_ttl" {
  hostname    = "api.example.com"
  record_type = "A"
  content     = "198.51.100.10"
  ttl         = 90
  enabled     = false
}
