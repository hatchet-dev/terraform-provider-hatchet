terraform {
  required_providers {
    hatchet = {
      source  = "hatchet-dev/hatchet"
      version = "~> 0.2.1"
    }
  }
}

provider "hatchet" {
  # Token is read from HATCHET_CLOUD_MANAGEMENT_TOKEN environment variable
}

# Create a new tenant
resource "hatchet_tenant" "production" {
  name = "Production Environment"
  slug = "production-123"
}

# Create another tenant
resource "hatchet_tenant" "staging" {
  name = "Staging Environment"
}