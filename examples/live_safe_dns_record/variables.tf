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

variable "test_ip" {
  description = "Safe documentation/test IPv4 value for the disposable A record."
  type        = string
  default     = "198.51.100.10"
}

variable "test_ttl" {
  description = "TTL in seconds for the disposable DNS record."
  type        = number
  default     = 300
}
