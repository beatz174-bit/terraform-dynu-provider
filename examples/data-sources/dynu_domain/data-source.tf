data "dynu_domain" "selected" {
  hostname = "www.example.com"
}

output "domain" {
  value = data.dynu_domain.selected.domain
}
