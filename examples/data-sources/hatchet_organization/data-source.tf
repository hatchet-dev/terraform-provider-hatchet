terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchet" {
  # Token is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
}

# Fetch organization information by ID
data "hatchet_organization" "example" {
  id = "17015348-d740-45f2-b23d-ea284c6eb3ee"
}
