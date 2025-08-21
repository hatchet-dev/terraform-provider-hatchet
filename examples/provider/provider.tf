terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # endpoint is optional and defaults to "cloud.onhatchet.run"
  endpoint = "cloud.onhatchet.run"

  # Management token for accessing the Hatchet Cloud API
  # This should be a sensitive value - consider using environment HATCHET_TOKEN
  # token = var.hatchet_management_token
}

# Example variable for the management token
variable "hatchet_management_token" {
  description = "Management token for accessing Hatchet Cloud API"
  type        = string
  sensitive   = true
}
