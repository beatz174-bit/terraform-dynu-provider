terraform {
  required_providers {
    dynu = {
      source = "dynu/dynu"
    }

    random = {
      source  = "hashicorp/random"
      version = ">= 3.6.0"
    }
  }
}

provider "dynu" {
  api_key = var.dynu_api_key
}

resource "random_id" "suffix" {
  byte_length = 4
}

locals {
  disposable_label    = "${var.test_prefix}-${lower(random_id.suffix.hex)}"
  disposable_hostname = "${local.disposable_label}.${var.dynu_root_domain}"
}

resource "dynu_dns_record" "safe_live_test" {
  hostname    = local.disposable_hostname
  record_type = "A"
  content     = var.test_ip
  ttl         = var.test_ttl
  state       = true
}
