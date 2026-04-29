variable "dynu_api_key" {
  description = "Dynu API key. Leave null to use DYNU_API_KEY from environment."
  type        = string
  default     = null
  sensitive   = true
}

variable "dynu_root_domain" {
  description = "Dynu-managed root domain used to create disposable live test records (for example: example.com)."
  type        = string

  validation {
    condition     = trimspace(var.dynu_root_domain) != ""
    error_message = "dynu_root_domain must be a non-empty Dynu-managed root domain."
  }
}

variable "test_suffix" {
  description = "Suffix used to make this run's disposable record hostnames unique."
  type        = string
  default     = "manual"

  validation {
    condition     = trimspace(var.test_suffix) != ""
    error_message = "test_suffix must be non-empty."
  }
}

variable "test_ipv4" {
  description = "IPv4 value for the disposable A record scenario."
  type        = string
  default     = "192.0.2.123"

  validation {
    condition     = trimspace(var.test_ipv4) != ""
    error_message = "test_ipv4 must be non-empty for the A record scenario."
  }
}

variable "test_ipv6" {
  description = "IPv6 value for the disposable AAAA record scenario."
  type        = string
  default     = "2001:db8::123"

  validation {
    condition     = trimspace(var.test_ipv6) != ""
    error_message = "test_ipv6 must be non-empty for the AAAA record scenario."
  }
}

variable "test_cname_target" {
  description = "CNAME target for the disposable CNAME record scenario."
  type        = string
  default     = "example.com"

  validation {
    condition     = trimspace(var.test_cname_target) != ""
    error_message = "test_cname_target must be non-empty for the CNAME scenario."
  }
}

variable "test_ttl" {
  description = "TTL in seconds for disposable DNS records. Must be 0 (provider/API default) or >= 90."
  type        = number
  default     = 300
}

variable "test_location" {
  description = "Optional Dynu location hint for A/AAAA records."
  type        = string
  default     = "us"
}
