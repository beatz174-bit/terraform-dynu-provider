resource "dynu_dns_record" "txt" {
  hostname    = "api.example.com"
  record_type = "TXT"
  content     = "managed-by-terraform"
  ttl         = 300
  state       = true
}
