output "disposable_hostname" {
  description = "Generated unique disposable hostname created by this run."
  value       = local.disposable_hostname
}

output "disposable_record_id" {
  description = "Dynu DNS record ID in domain_id/record_id format for the disposable record."
  value       = dynu_dns_record.safe_live_test.id
}

output "root_domain" {
  description = "Dynu root domain used for this live-safe test."
  value       = var.dynu_root_domain
}

output "cleanup_reminder" {
  description = "Reminder to destroy the disposable record after verification."
  value       = "Run terraform destroy in this same directory/state to remove only this disposable record."
}
