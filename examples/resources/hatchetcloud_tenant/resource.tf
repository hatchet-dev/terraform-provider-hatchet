terraform {
  required_providers {
    hatchetcloud = {
      source  = "hatchet-dev/hatchetcloud"
      version = "~> 0.1.0"
    }
  }
}

provider "hatchetcloud" {
  # Token can be provided via HATCHET_TOKEN environment variable
}

# Reference an existing organization
data "hatchetcloud_organization" "existing" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Create a new tenant
resource "hatchetcloud_tenant" "production" {
  organization_id = data.hatchetcloud_organization.existing.id
  name            = "Production Environment"
  slug            = "prod"
}

# Create multiple tenants using for_each
variable "environments" {
  description = "Map of environments to create"
  type = map(object({
    name = string
    slug = string
  }))
  default = {
    "staging" = {
      name = "Staging Environment"
      slug = "staging"
    }
    "development" = {
      name = "Development Environment"
      slug = "dev"
    }
  }
}

resource "hatchetcloud_tenant" "environments" {
  for_each = var.environments

  organization_id = data.hatchetcloud_organization.existing.id
  name            = each.value.name
  slug            = each.value.slug
}

# Output tenant information
output "production_tenant_id" {
  description = "The ID of the production tenant"
  value       = hatchetcloud_tenant.production.id
}

output "all_tenant_ids" {
  description = "Map of all created tenant IDs"
  value       = { for k, v in hatchetcloud_tenant.environments : k => v.id }
}