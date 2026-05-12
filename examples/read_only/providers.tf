terraform {
  required_version = ">= 1.5.0"

  required_providers {
    dynu = {
      source  = "beatz174-bit/dynu"
      version = "~> 0.1.0"
    }
  }
}

# Local development note:
# This source address stays "beatz174-bit/dynu" while using ~/.terraformrc dev_overrides.
# Terraform will load your local terraform-provider-dynu binary instead of the public registry.
provider "dynu" {
  # Set var.dynu_api_key in terraform.tfvars.
  api_key = var.dynu_api_key
}
