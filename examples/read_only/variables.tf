variable "dynu_api_key" {
  description = "Dynu API key. Leave null to use DYNU_API_KEY from environment."
  type        = string
  default     = null
  sensitive   = true
}

variable "hostname" {
  description = "Fully-qualified hostname used by dynu_domain and dynu_dns_records data sources."
  type        = string
  default     = "www.example.com"
}
