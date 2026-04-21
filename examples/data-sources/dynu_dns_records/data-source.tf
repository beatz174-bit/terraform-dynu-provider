terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {}

data "dynu_dns_records" "records" {
  hostname = "www.example.com"
}

output "records" {
  value = data.dynu_dns_records.records.records
}
