terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {
  api_key = var.dynu_api_key
}

data "dynu_dns_records" "selected" {
  hostname = var.hostname
}
