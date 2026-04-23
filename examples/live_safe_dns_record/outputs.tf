output "hostname" {
  description = "Hostname queried by this read-only live-safe example."
  value       = var.hostname
}

output "records" {
  description = "DNS records returned for the selected hostname."
  value       = data.dynu_dns_records.selected.records
}

output "record_count" {
  description = "Number of DNS records returned for the selected hostname."
  value       = length(data.dynu_dns_records.selected.records)
}
