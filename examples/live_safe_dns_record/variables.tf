variable "dynu_api_key" {
  description = "Dynu API key. Leave null to use DYNU_API_KEY from the environment."
  type        = string
  default     = null
  sensitive   = true
}

variable "dynu_root_domain" {
  description = "Dynu-managed root domain for creating a disposable live test subdomain (for example: example.com)."
  type        = string
}

variable "test_prefix" {
  description = "Prefix for the disposable test hostname label."
  type        = string
  default     = "tfacc"
}

variable "test_suffix" {
  description = "Suffix for disposable hostname uniqueness (set explicitly per run if needed)."
  type        = string
  default     = "manual"
}

variable "test_ip" {
  description = "IPv4 value for the disposable A record. Dynu rejects reserved documentation ranges (for example 198.51.100.0/24), so this default uses a public resolver address."
  type        = string
  default     = "1.1.1.1"
}

variable "test_ttl" {
  description = "TTL in seconds for the disposable DNS record."
  type        = number
  default     = 300
}
