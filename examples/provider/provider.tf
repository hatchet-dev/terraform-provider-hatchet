terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 1.0"
    }
  }
}

provider "hatchetcloud" {
  # endpoint is optional and defaults to "cloud.onhatchet.run"
  # token is required
  token = "your-api-token-here"
}
