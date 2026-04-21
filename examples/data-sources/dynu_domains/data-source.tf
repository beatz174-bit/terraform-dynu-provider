data "dynu_domains" "all" {}

output "domains" {
  value = data.dynu_domains.all.domains
}
