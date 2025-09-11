terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 0.2.1"
    }
  }
}

provider "hatchet" {
  # JWT is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
  # or specify it here (not recommended for production)
  # token = "eyJhbGciOiJFUzI1NiIs..."
}