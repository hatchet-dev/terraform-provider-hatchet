terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # JWT token is read from HATCHET_TOKEN environment variable
  # or specify it here (not recommended for production)
  # token = "eyJhbGciOiJFUzI1NiIs..."
}