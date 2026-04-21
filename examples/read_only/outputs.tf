output "domains" {
  description = "All visible Dynu domains."
  value       = data.dynu_domains.all.domains
  sensitive   = true
}

output "resolved_domain" {
  description = "Resolved Dynu root domain details for var.hostname."
  value       = data.dynu_domain.selected.domain
  sensitive   = true
}

output "resolved_dns_records" {
  description = "DNS records for the resolved root domain."
  value       = data.dynu_dns_records.selected.records
}

output "resolved_domain_id" {
  description = "Domain ID returned by dynu_dns_records."
  value       = data.dynu_dns_records.selected.domain_id
}

output "resolved_domain_name" {
  description = "Root domain name returned by dynu_dns_records."
  value       = data.dynu_dns_records.selected.domain_name
}
