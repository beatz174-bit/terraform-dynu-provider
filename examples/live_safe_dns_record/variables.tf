variable "dynu_api_key" {
  description = "Dynu API key. Leave null to use DYNU_API_KEY from the environment."
  type        = string
  default     = null
  sensitive   = true
}

variable "hostname" {
  description = "Existing Dynu-managed hostname to inspect (for example: www.example.com)."
  type        = string
}
