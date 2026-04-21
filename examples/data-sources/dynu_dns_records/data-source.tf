data "dynu_dns_records" "records" {
  hostname = "www.example.com"
}

output "records" {
  value = data.dynu_dns_records.records.records
}
