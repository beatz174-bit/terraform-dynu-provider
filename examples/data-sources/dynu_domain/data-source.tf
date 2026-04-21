terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {}

data "dynu_domain" "selected" {
  hostname = "www.example.com"
}

output "domain" {
  value = data.dynu_domain.selected.domain
}
