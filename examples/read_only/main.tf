# Lists all Dynu domains available to the configured API key.
data "dynu_domains" "all" {}

# Resolves the Dynu root domain for a specific hostname.
data "dynu_domain" "selected" {
  hostname = var.hostname
}

# Lists DNS records for the root domain resolved from the same hostname.
data "dynu_dns_records" "selected" {
  hostname = var.hostname
}
