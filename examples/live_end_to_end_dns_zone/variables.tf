variable "dynu_api_key" {
  description = "Dynu API key. Leave null to use DYNU_API_KEY from environment."
  type        = string
  default     = null
  sensitive   = true
}

variable "test_domain" {
  type        = string
  description = "Disposable Dynu test domain to create and destroy during live end-to-end testing."

  validation {
    condition     = trimspace(var.test_domain) != ""
    error_message = "test_domain must be a non-empty domain name."
  }
}

variable "test_ttl" {
  description = "TTL in seconds for disposable DNS records. Must be 0 (provider/API default) or >= 90."
  type        = number
  default     = 300
}
