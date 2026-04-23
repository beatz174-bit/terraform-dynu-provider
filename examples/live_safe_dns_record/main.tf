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

locals {
  disposable_label    = "${var.test_prefix}-${var.test_suffix}"
  disposable_hostname = "${local.disposable_label}.${var.dynu_root_domain}"
}

resource "dynu_dns_record" "safe_live_test" {
  hostname    = local.disposable_hostname
  record_type = "A"
  content     = var.test_ip
  ttl         = var.test_ttl
  state       = true
}
