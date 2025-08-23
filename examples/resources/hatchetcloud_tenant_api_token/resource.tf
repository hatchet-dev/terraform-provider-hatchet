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

# Create a tenant first
resource "hatchetcloud_tenant" "example" {
  name = "Example Tenant"
  slug = "example"
}

# Create an API token for the tenant
resource "hatchetcloud_tenant_api_token" "production_token" {
  tenant_id = hatchetcloud_tenant.example.id
  name      = "Production API Token"
}

# Create a temporary API token with expiration
resource "hatchetcloud_tenant_api_token" "temp_token" {
  tenant_id  = hatchetcloud_tenant.example.id
  name       = "Temporary Access Token"
  expires_at = "24h"
}