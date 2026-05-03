variable "dynu_api_key" {
  description = "Dynu API key used by the provider. Set in terraform.tfvars."
  type        = string
  default     = null
  sensitive   = true
}

variable "hostname" {
  description = "Fully-qualified hostname used by dynu_domain and dynu_dns_records data sources."
  type        = string
  default     = "www.example.com"
}
