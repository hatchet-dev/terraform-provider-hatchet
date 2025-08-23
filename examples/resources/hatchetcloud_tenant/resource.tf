terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # Token is read from HATCHET_TOKEN environment variable
}

# Create a new tenant
resource "hatchetcloud_tenant" "production" {
  name = "Production Environment"
  slug = "prod"
}

# Create another tenant
resource "hatchetcloud_tenant" "staging" {
  name = "Staging Environment"
  slug = "staging"
}