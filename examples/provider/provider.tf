terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }
  }
}

provider "dynu" {
  # api_key can be omitted when DYNU_API_KEY is set.
  api_key = var.dynu_api_key
}

variable "dynu_api_key" {
  type      = string
  sensitive = true
}
