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

# Fetch tenant information by ID
data "hatchetcloud_tenant" "production" {
  id = "87654321-4321-4321-4321-210987654321"
}

# Use tenant data to create API tokens
resource "hatchetcloud_tenant_api_token" "production_api_token" {
  tenant_id = data.hatchetcloud_tenant.production.id
  name      = "API Token for Production Tenant"
}