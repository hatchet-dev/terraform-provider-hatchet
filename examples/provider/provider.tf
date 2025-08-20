terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 1.0"
    }
  }
}

provider "hatchetcloud" {
  endpoint = "https://api.hatchet.cloud"
  token    = "your-api-token-here"
}
