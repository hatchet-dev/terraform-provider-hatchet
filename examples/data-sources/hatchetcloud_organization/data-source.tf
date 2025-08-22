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

# Fetch organization information by ID
data "hatchetcloud_organization" "example" {
  id = "12345678-1234-1234-1234-123456789012"
}

# Use the organization data in other resources
resource "hatchetcloud_tenant" "new_tenant" {
  org_id = data.hatchetcloud_organization.example.id
  name   = "Tenant for ${data.hatchetcloud_organization.example.name}"
  slug   = "new-tenant"
}

# Use organization data for conditional logic
variable "target_organization_name" {
  description = "Name of the organization to match"
  type        = string
  default     = "Production Org"
}

locals {
  is_target_org = data.hatchetcloud_organization.example.name == var.target_organization_name
}

resource "hatchetcloud_tenant" "conditional_tenant" {
  count = local.is_target_org ? 1 : 0

  org_id = data.hatchetcloud_organization.example.id
  name   = "Conditional Tenant"
  slug   = "conditional"
}

# Output organization information
output "organization_name" {
  description = "The name of the organization"
  value       = data.hatchetcloud_organization.example.name
}

output "org_id" {
  description = "The ID of the organization"
  value       = data.hatchetcloud_organization.example.id
}

output "tenant_created" {
  description = "Whether the conditional tenant was created"
  value       = local.is_target_org
}