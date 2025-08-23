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

# Fetch organization information by ID
data "hatchetcloud_organization" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Use the organization data to create a tenant
resource "hatchetcloud_tenant" "new_tenant" {
  name = "Tenant for ${data.hatchetcloud_organization.example.name}"
  slug = "new-tenant"
}