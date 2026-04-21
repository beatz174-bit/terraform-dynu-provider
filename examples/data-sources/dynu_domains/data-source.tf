terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {}

data "dynu_domains" "all" {}

output "domains" {
  value = data.dynu_domains.all.domains
}
